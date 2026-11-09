package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const CallbackPlayersPrefix = "players:"

func (h *Handlers) playerCountsReply(ctx context.Context, chatID int64, category, optionID string) tgbotapi.MessageConfig {
	counts, err := h.matchmaking.PlayerCounts(ctx, chatID)
	if err != nil || len(counts) == 0 {
		return tgbotapi.NewMessage(chatID, "⚠️ Player counts are unavailable. Please try /play again shortly.")
	}
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for i, count := range counts {
		if i%3 == 0 {
			rows = append(rows, []tgbotapi.InlineKeyboardButton{})
		}
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d players", count), fmt.Sprintf("%s%s|%s|%d", CallbackPlayersPrefix, category, optionID, count))
		rows[len(rows)-1] = append(rows[len(rows)-1], button)
	}
	reply := tgbotapi.NewMessage(chatID, "👥 How many players should be in your game, including you?")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return reply
}

func (h *Handlers) PlayerCountCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.Chattable, error) {
	parts := strings.Split(strings.TrimPrefix(cb.Data, CallbackPlayersPrefix), "|")
	if len(parts) != 3 || (parts[0] != "SPORT" && parts[0] != "MUSIC" && parts[0] != "TECH") {
		return tgbotapi.NewMessage(cb.Message.Chat.ID, "⚠️ Invalid selection. Please use /play."), nil
	}
	players, err := strconv.Atoi(parts[2])
	if err != nil || players < 2 {
		return tgbotapi.NewMessage(cb.Message.Chat.ID, "⚠️ Invalid player count. Please use /play."), nil
	}
	// Recheck the timer too: an old keyboard must not revive a disabled option.
	options, err := h.matchmaking.TimeOptions(ctx, cb.Message.Chat.ID)
	if err != nil {
		return tgbotapi.NewMessage(cb.Message.Chat.ID, "⚠️ Could not load question times. Please use /play again."), nil
	}
	for _, option := range options {
		if option.ID == parts[1] && option.Enabled {
			return h.joinCategory(ctx, cb, parts[0], parts[1], players), nil
		}
	}
	return h.timeOptionsReply(ctx, cb.Message.Chat.ID, parts[0]), nil
}
