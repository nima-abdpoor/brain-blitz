package commands

import (
	"context"
	"errors"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/matchmaking"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/telegram/keyboards"
)

// CategoryCallback handles a category button from /play's keyboard (see
// keyboards.Category). internal/telegram/middleware.RequireAuthCallback
// gates this callback (cmd/bot/main.go), which is registered once per
// internal/core/matchmaking.Categories entry since Bot.RegisterCallback
// matches exact data, not a prefix.
func (h *Handlers) CategoryCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.Chattable, error) {
	chatID := cb.Message.Chat.ID
	category := strings.TrimPrefix(cb.Data, keyboards.CallbackCategoryPrefix)

	if cb.From != nil {
		h.matchmaking.SetDisplayName(ctx, chatID, strings.TrimSpace(cb.From.FirstName+" "+cb.From.LastName))
	}
	return h.timeOptionsReply(ctx, chatID, category), nil
}

func (h *Handlers) joinCategory(ctx context.Context, cb *tgbotapi.CallbackQuery, category, optionID string) tgbotapi.Chattable {
	chatID := cb.Message.Chat.ID
	err := h.matchmaking.JoinQueue(ctx, chatID, category, optionID)
	switch {
	case err == nil:
		reply := tgbotapi.NewMessage(chatID, "🔎 Waiting for opponent…")
		reply.ReplyMarkup = keyboards.Waiting()
		return reply
	case errors.Is(err, matchmaking.ErrAlreadyInProgress):
		return tgbotapi.NewMessage(chatID, "You're already looking for a match.")
	case errors.Is(err, matchmaking.ErrSessionExpired):
		reply := tgbotapi.NewMessage(chatID, "Your session has expired. Send /start to log in again.")
		reply.ReplyMarkup = keyboards.Auth()
		return reply
	default:
		return tgbotapi.NewMessage(chatID, "Couldn't connect to the game right now. Please try /play again shortly.")
	}
}

// CancelCallback closes the waiting socket; the backend removes its queue membership.
func (h *Handlers) CancelCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.Chattable, error) {
	chatID := cb.Message.Chat.ID

	if err := h.matchmaking.Cancel(chatID); err != nil {
		reply := tgbotapi.NewMessage(chatID, "You're not waiting for a match right now.")
		reply.ReplyMarkup = keyboards.MainMenu()
		return reply, nil
	}

	reply := tgbotapi.NewMessage(chatID,
		"❌ You've stopped waiting. Choose Play when you're ready for another match.")
	reply.ReplyMarkup = keyboards.MainMenu()
	return reply, nil
}
