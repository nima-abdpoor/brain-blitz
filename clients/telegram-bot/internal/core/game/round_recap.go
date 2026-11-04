package game

import (
	"encoding/json"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/ws"
)

type RoundRecap struct {
	QuestionID    string `json:"questionId"`
	Question      string `json:"question"`
	PlayerAnswer  string `json:"playerAnswer"`
	CorrectAnswer string `json:"correctAnswer"`
	Answered      bool   `json:"answered"`
	IsCorrect     bool   `json:"isCorrect"`
}
type roundNotifier interface{ NotifyRoundClosed(int64, RoundRecap, int) }

func (m *Manager) handleRoundClosed(chatID int64, cg *chatGame, event ws.Event) {
	var meta struct {
		GameID      string      `json:"gameId"`
		RoundNumber int         `json:"roundNumber"`
		Recap       *RoundRecap `json:"recap"`
	}
	if json.Unmarshal(event.MetaData, &meta) != nil || meta.Recap == nil {
		return
	}
	m.mu.Lock()
	if m.chats[chatID] != cg || meta.GameID != cg.gameID || meta.RoundNumber != cg.roundNumber || meta.RoundNumber <= cg.lastRecapRound {
		m.mu.Unlock()
		return
	}
	cg.lastRecapRound = meta.RoundNumber
	m.mu.Unlock()
	if notifier, ok := m.notifier.(roundNotifier); ok {
		notifier.NotifyRoundClosed(chatID, *meta.Recap, meta.RoundNumber)
	}
}
