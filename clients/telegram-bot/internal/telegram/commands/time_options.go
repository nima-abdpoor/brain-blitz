package commands

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

const CallbackTimePrefix = "time:"

func (h *Handlers) timeOptionsReply(ctx context.Context, chatID int64, category string) tgbotapi.MessageConfig {
	options, err := h.matchmaking.TimeOptions(ctx, chatID)
	if err != nil || len(options) == 0 {
		return tgbotapi.NewMessage(chatID, "⚠️ Question time options are unavailable. Please try /play again shortly.")
	}
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, o := range options {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⏱️ "+o.Label, CallbackTimePrefix+category+"|"+o.ID)))
	}
	reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("⏱️ How much time should each %s question allow?", strings.ToLower(category)))
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return reply
}
func (h *Handlers) TimeOptionCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.Chattable, error) {
	category, id, ok := strings.Cut(strings.TrimPrefix(cb.Data, CallbackTimePrefix), "|")
	if !ok || (category != "SPORT" && category != "MUSIC" && category != "TECH") {
		return tgbotapi.NewMessage(cb.Message.Chat.ID, "⚠️ Invalid selection. Please use /play."), nil
	}
	options, err := h.matchmaking.TimeOptions(ctx, cb.Message.Chat.ID)
	if err != nil {
		return tgbotapi.NewMessage(cb.Message.Chat.ID, "⚠️ Could not load question times. Please use /play again."), nil
	}
	found := false
	for _, o := range options {
		found = found || (o.ID == id && o.Enabled)
	}
	if !found {
		return h.timeOptionsReply(ctx, cb.Message.Chat.ID, category), nil
	}
	reply := h.playerCountsReply(ctx, cb.Message.Chat.ID, category, id)
	return reply, nil
}
