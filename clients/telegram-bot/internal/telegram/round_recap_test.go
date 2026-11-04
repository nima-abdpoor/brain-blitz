package telegram

import (
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
	"strings"
	"testing"
)

func TestFormatRoundRecap(t *testing.T) {
	// Exercise the same formatter used by live notifications, including timeouts.
	for _, recap := range []game.RoundRecap{{Question: "Sport?", PlayerAnswer: "Archery", CorrectAnswer: "Archery", Answered: true, IsCorrect: true}, {Question: "Sport?", PlayerAnswer: "Fencing", CorrectAnswer: "Archery", Answered: true}, {Question: "Sport?", CorrectAnswer: "Archery"}} {
		text := formatRoundRecap(recap, 1)
		if !strings.Contains(text, "Correct answer: Archery") {
			t.Fatal(text)
		}
		if !recap.Answered && !strings.Contains(text, "No answer submitted") {
			t.Fatal(text)
		}
	}
}
