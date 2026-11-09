package commands

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/matchmaking"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
)

type countsAPI struct{ counts []int }

func (a countsAPI) PlayerCounts(context.Context, string) ([]int, error) { return a.counts, nil }

func playerCountHandlers(t *testing.T) *Handlers {
	t.Helper()
	sessions := session.NewMemoryStore()
	if err := sessions.Save(context.Background(), session.Session{TelegramChatID: 1, AccessToken: "at"}); err != nil {
		t.Fatal(err)
	}
	mgr := newTestMatchmakingManager(sessions)
	mgr.SetTimeOptionsAPI(optionsAPI{options: []matchmaking.TimeOption{{ID: "fast", Enabled: true, Seconds: 10}}})
	return &Handlers{matchmaking: mgr}
}

func TestTimerSelectionPromptsForConfiguredPlayerCounts(t *testing.T) {
	h := playerCountHandlers(t)
	counts := []int{}
	for i := 2; i <= 15; i++ {
		counts = append(counts, i)
	}
	h.matchmaking.SetPlayerCountsAPI(countsAPI{counts})
	cb := &tgbotapi.CallbackQuery{Data: "time:TECH|fast", Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}}}
	response, err := h.TimeOptionCallback(context.Background(), cb)
	if err != nil {
		t.Fatal(err)
	}
	msg := response.(tgbotapi.MessageConfig)
	if !strings.Contains(msg.Text, "including you") {
		t.Fatal(msg.Text)
	}
	kb := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	var buttons []tgbotapi.InlineKeyboardButton
	for _, row := range kb.InlineKeyboard {
		buttons = append(buttons, row...)
	}
	if len(buttons) != 14 {
		t.Fatalf("got %d buttons", len(buttons))
	}
	for i, button := range buttons {
		if *button.CallbackData != fmt.Sprintf("players:TECH|fast|%d", i+2) || len(*button.CallbackData) > 64 {
			t.Fatal(button)
		}
	}
	if h.matchmaking.StateFor(1) != matchmaking.Idle {
		t.Fatal("timer selection joined before player choice")
	}
	h.matchmaking.SetPlayerCountsAPI(countsAPI{[]int{3, 4}})
	msg = h.playerCountsReply(context.Background(), 1, "TECH", "fast")
	kb = msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if len(kb.InlineKeyboard) != 1 || len(kb.InlineKeyboard[0]) != 2 {
		t.Fatal("ignored configured limits")
	}
}

func TestInvalidAndUnavailablePlayerSelectionsDoNotQueue(t *testing.T) {
	h := playerCountHandlers(t)
	h.matchmaking.SetPlayerCountsAPI(countsAPI{[]int{3, 4}})
	for _, data := range []string{"players:TECH|fast|1", "players:TECH|fast|no", "players:TECH|fast|15", "players:INVALID|fast|3", "players:TECH|fast", "players:TECH|disabled|3"} {
		cb := &tgbotapi.CallbackQuery{Data: data, Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: 1}}}
		if _, err := h.PlayerCountCallback(context.Background(), cb); err != nil {
			t.Fatal(err)
		}
		if h.matchmaking.StateFor(1) != matchmaking.Idle {
			t.Fatalf("queued invalid selection %q", data)
		}
	}
	h.matchmaking.SetPlayerCountsAPI(nil)
	msg := h.playerCountsReply(context.Background(), 1, "TECH", "fast")
	if !strings.Contains(msg.Text, "unavailable") {
		t.Fatal(msg.Text)
	}
}
