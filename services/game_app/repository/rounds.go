package repository

import (
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"fmt"
	"time"
)

func (m GameRepository) OpenRound(ctx context.Context, gameID string, index int) (service.GameQuestions, error) {
	q, err := m.GetQuestionsByGameId(ctx, gameID)
	if err != nil {
		return q, err
	}
	if index < 0 || index > len(q.Questions) || m.answerWindow(q) <= 0 {
		return q, fmt.Errorf("invalid round or answer timeout")
	}
	q.RoundBased = true
	q.CurrentQuestionIndex = index
	q.Completed = index == len(q.Questions)
	if !q.Completed {
		q.Questions[index].ValidAnswerTime = time.Now().UTC().Add(m.answerWindow(q))
	}
	return q, m.saveQuestionByGameId(ctx, gameID, q)
}

func (m GameRepository) answerWindow(q service.GameQuestions) time.Duration {
	if q.QuestionSeconds > 0 {
		return time.Duration(q.QuestionSeconds) * time.Second
	}
	return m.Config.ValidAnswerTimeOut
}
func (m GameRepository) calculateRoundScore(correct bool, answer, deadline time.Time, window time.Duration) int {
	if window > 0 && m.Config.ValidAnswerTimeOut > 0 {
		m.Config.ScoreConfig.BonusDeadline = time.Duration(float64(m.Config.ScoreConfig.BonusDeadline) * float64(window) / float64(m.Config.ValidAnswerTimeOut))
	}
	return m.CalculateScore(correct, answer, deadline)
}
