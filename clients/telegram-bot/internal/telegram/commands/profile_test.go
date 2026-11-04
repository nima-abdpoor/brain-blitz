package commands

import (
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/profile"
	"strings"
	"testing"
)

func TestFormatStatistics(t *testing.T) {
	text := formatStatistics(&profile.Statistics{GamesPlayed: 3, Wins: 1, Losses: 1, Draws: 1, TotalPoints: 44, BestScore: 30, AveragePoints: 14.7, Rank: 2, RankedPlayers: 10, CorrectAnswers: 5, AnswersSubmitted: 10, Accuracy: 50})
	for _, want := range []string{"Games played: 3", "Wins: 1 | Losses: 1 | Draws: 1", "Total points: 44", "Average score: 14.7 | Best score: 30", "5/10 (50.0% accuracy)", "#2 of 10 players"} {
		if !strings.Contains(text, want) {
			t.Errorf("profile missing %q: %s", want, text)
		}
	}
	if !strings.Contains(formatStatistics(&profile.Statistics{}), "Unranked") {
		t.Fatal("new player must be unranked")
	}
	if !strings.Contains(formatStatistics(nil), "unavailable") {
		t.Fatal("missing stats must not show zero totals")
	}
}
