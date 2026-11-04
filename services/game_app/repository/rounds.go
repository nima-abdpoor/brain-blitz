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
	if index < 0 || index > len(q.Questions) || m.Config.ValidAnswerTimeOut <= 0 {
		return q, fmt.Errorf("invalid round or answer timeout")
	}
	q.RoundBased = true
	q.CurrentQuestionIndex = index
	q.Completed = index == len(q.Questions)
	if !q.Completed {
		q.Questions[index].ValidAnswerTime = time.Now().UTC().Add(m.Config.ValidAnswerTimeOut)
	}
	return q, m.saveQuestionByGameId(ctx, gameID, q)
}
