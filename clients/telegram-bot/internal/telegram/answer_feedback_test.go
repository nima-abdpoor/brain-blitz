package telegram

import (
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
)

func TestAnswerAcceptedShowsServerVerdictWithScore(t *testing.T) {
	for _, tc := range []struct {
		correct bool
		text    string
	}{
		{correct: true, text: "✅ Correct!"},
		{correct: false, text: "❌ Incorrect."},
	} {
		t.Run(tc.text, func(t *testing.T) {
			sender := &fakeSender{}
			n := NewGameNotifier(sender, testLogger())
			n.NotifyAnswerAccepted(42, tc.correct, game.Leaderboard{Scores: []game.PlayerScore{{DisplayName: "Alex", Points: 0}}})
			if len(sender.sent) != 1 {
				t.Fatalf("sent %d notifications, want 1", len(sender.sent))
			}
			msg := sender.sent[0].(tgbotapi.MessageConfig)
			if msg.ChatID != 42 || !strings.Contains(msg.Text, tc.text) || !strings.Contains(msg.Text, "Alex: 0 pts") || !strings.Contains(msg.Text, "everyone answers or time expires") {
				t.Fatalf("wrong feedback: %+v", msg)
			}
			if strings.Contains(msg.Text, "Correct answer:") {
				t.Fatal("feedback must not reveal the solution before round closure")
			}
		})
	}
}
