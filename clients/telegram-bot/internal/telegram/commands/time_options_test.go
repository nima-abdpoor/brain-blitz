package commands

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/matchmaking"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
	"strings"
	"testing"
)

type optionsAPI struct{ options []matchmaking.TimeOption }

func (a optionsAPI) TimeOptions(context.Context, string) ([]matchmaking.TimeOption, error) {
	return a.options, nil
}
func TestCategoryPromptsForDatabaseOptions(t *testing.T) {
	sessions := session.NewMemoryStore()
	sessions.Save(context.Background(), session.Session{TelegramChatID: 1, AccessToken: "at"})
	mgr := newTestMatchmakingManager(sessions)
	mgr.SetTimeOptionsAPI(optionsAPI{[]matchmaking.TimeOption{{ID: "10-seconds", Label: "10 seconds", Seconds: 10, Enabled: true}, {ID: "1-minute", Label: "1 minute", Seconds: 60, Enabled: true}}})
	h := &Handlers{matchmaking: mgr}
	cb := &tgbotapi.CallbackQuery{Data: "category:SPORT", Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}}}
	response, err := h.CategoryCallback(context.Background(), cb)
	if err != nil {
		t.Fatal(err)
	}
	msg := response.(tgbotapi.MessageConfig)
	if !strings.Contains(msg.Text, "How much time") {
		t.Fatal(msg.Text)
	}
	kb := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if len(kb.InlineKeyboard) != 2 || *kb.InlineKeyboard[0][0].CallbackData != "time:SPORT|10-seconds" {
		t.Fatal("missing database choices")
	}
	if mgr.StateFor(1) != matchmaking.Idle {
		t.Fatal("category selection must not join before timer choice")
	}
}
