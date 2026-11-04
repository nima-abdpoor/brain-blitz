package telegram

import (
	"context"
	"strings"
	"testing"

	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
)

func TestResultsUseIdentityAndNamesRatherThanBoardOrder(t *testing.T) {
	sessions := session.NewMemoryStore()
	_ = sessions.Save(context.Background(), session.Session{TelegramChatID: 100, UserID: "13"})
	_ = sessions.Save(context.Background(), session.Session{TelegramChatID: 200, UserID: "14"})
	n := NewGameNotifier(&fakeSender{}, testLogger(), sessions)
	b := game.Leaderboard{Scores: []game.PlayerScore{{PlayerID: "14", DisplayName: "Sam", Points: 45}, {PlayerID: "13", DisplayName: "Alex", Points: 50}}}
	if !strings.Contains(n.outcome(100, b), "You won!") || !strings.Contains(n.outcome(200, b), "Good game!") {
		t.Fatal("wrong player outcome")
	}
	text := formatLeaderboard("Results", b)
	if !strings.Contains(text, "Alex: 50") || strings.Contains(text, "Player 13") {
		t.Fatal(text)
	}
	b.Scores[0].Points = 50
	if !strings.Contains(n.outcome(100, b), "draw") || !strings.Contains(n.outcome(200, b), "draw") {
		t.Fatal("tied scores must draw")
	}
}
