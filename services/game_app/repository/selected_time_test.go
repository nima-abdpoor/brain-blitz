package repository

import (
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSelectedQuestionWindowAndProportionalScore(t *testing.T) {
	repo := newTestRepo(t)
	for _, seconds := range []int{10, 60} {
		q := service.GameQuestions{QuestionSeconds: seconds, Questions: []service.Question{{Id: "q"}}}
		require.NoError(t, repo.saveQuestionByGameId(context.Background(), "selected", q))
		start := time.Now()
		opened, err := repo.OpenRound(context.Background(), "selected", 0)
		require.NoError(t, err)
		require.WithinDuration(t, start.Add(time.Duration(seconds)*time.Second), opened.Questions[0].ValidAnswerTime, 100*time.Millisecond)
		deadline := time.Now().Add(time.Duration(seconds) * time.Second)
		score := repo.calculateRoundScore(true, deadline.Add(-time.Duration(seconds)*time.Second/2), deadline, time.Duration(seconds)*time.Second)
		require.Greater(t, score, repo.Config.ScoreConfig.BaseScore)
		require.LessOrEqual(t, score, repo.Config.ScoreConfig.BaseScore+repo.Config.ScoreConfig.MaxBonus)
	}
}
