package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/game"
)

func (n *GameNotifier) NotifyRoundClosed(chatID int64, recap game.RoundRecap, round int) {
	n.stopCountdown(chatID)
	n.send(chatID, tgbotapi.NewMessage(chatID, formatRoundRecap(recap, round)))
}
func formatRoundRecap(recap game.RoundRecap, round int) string {
	answer := "No answer submitted"
	result := "⏳ Time expired"
	if recap.Answered {
		answer = recap.PlayerAnswer
		result = "❌ Incorrect"
		if recap.IsCorrect {
			result = "✅ Correct!"
		}
	}
	text := fmt.Sprintf("📝 Question %d result\n%s\n\nYour answer: %s\nCorrect answer: %s\n%s", round, recap.Question, answer, recap.CorrectAnswer, result)
	return text
}
