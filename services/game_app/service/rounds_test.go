package service

import (
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type roundRepository struct {
	Repository
	q       GameQuestions
	answers map[string][]QuestionCorrectness
	status  GameStatus
	opens   int
}

func (r *roundRepository) GetQuestionsByGameId(context.Context, string) (GameQuestions, error) {
	return r.q, nil
}
func (r *roundRepository) OpenRound(_ context.Context, _ string, i int) (GameQuestions, error) {
	r.opens++
	r.q.RoundBased = true
	r.q.CurrentQuestionIndex = i
	r.q.Completed = i == len(r.q.Questions)
	if !r.q.Completed {
		r.q.Questions[i].ValidAnswerTime = time.Now().Add(time.Second)
	}
	return r.q, nil
}
func (r *roundRepository) SavePlayerAnswer(_ context.Context, _ string, a PlayerAnswer) error {
	for _, v := range r.answers[a.PlayerID] {
		if v.QuestionId == a.QuestionIDs {
			return fmt.Errorf("duplicate")
		}
	}
	r.answers[a.PlayerID] = append(r.answers[a.PlayerID], QuestionCorrectness{QuestionId: a.QuestionIDs})
	return nil
}
func (r *roundRepository) GetLeaderBoard(context.Context, string) (LeaderBoard, error) {
	b := LeaderBoard{}
	for _, id := range []string{"1", "2"} {
		b.PlayersPoint = append(b.PlayersPoint, PlayerPoint{PlayerId: id, QuestionCorrectness: r.answers[id]})
	}
	return b, nil
}
func (r *roundRepository) GetGame(context.Context, string) (Game, error) {
	return Game{Players: []uint64{1, 2}, Status: r.status}, nil
}
func (r *roundRepository) FinishGame(context.Context, string) error {
	r.status = GameStatusFinished
	return nil
}
func (r *roundRepository) UpsertUserStatus(context.Context, uint64, GameStatus) error { return nil }

type roundPublisher struct{}

func (r *roundRepository) SetValidAnswerTimeForQuestions(context.Context, string) (time.Duration, error) {
	return time.Second, nil
}

func (roundPublisher) Publish(context.Context, string, any, ...taskqueue.Option) (string, error) {
	return "task", nil
}

func TestConcurrentAnswersAdvanceOnceAndTimeoutDoesNotSkip(t *testing.T) {
	r := &roundRepository{q: GameQuestions{Players: []uint64{1, 2}, Questions: []Question{{Id: "q1"}, {Id: "q2"}}}, answers: map[string][]QuestionCorrectness{}}
	sock := &completionSocket{}
	svc := NewService(Config{}, r, sock, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	ctx := context.Background()
	require.NoError(t, svc.startRounds(ctx, "g"))
	deadline := r.q.Questions[0].ValidAnswerTime
	require.NoError(t, svc.startRounds(ctx, "g"))
	require.Equal(t, 1, r.opens)
	require.Equal(t, deadline, r.q.Questions[0].ValidAnswerTime)
	var wg sync.WaitGroup
	for _, id := range []uint64{1, 2, 1, 2} {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			var conn net.Conn
			msg, _ := json.Marshal(ProcessGameMessageRequest{Command: CommandAnswer, GameAnswer: ProcessGameAnswer{GameId: "g", QuestionId: "q1", Answer: "wrong"}})
			_ = svc.readMessage(ctx, id, &conn, websocket.OpText, string(msg))
		}(id)
	}
	wg.Wait()
	require.Equal(t, 1, r.q.CurrentQuestionIndex)
	require.Equal(t, 2, r.opens)
	require.NoError(t, svc.ProcessRoundTimeout(ctx, map[string]interface{}{"gameId": "g", "round": float64(0)}))
	require.Equal(t, 1, r.q.CurrentQuestionIndex)
	r.q.Questions[1].ValidAnswerTime = time.Now().Add(-time.Second)
	require.NoError(t, svc.ProcessRoundTimeout(ctx, map[string]interface{}{"gameId": "g", "round": float64(1)}))
	require.Equal(t, GameStatusFinished, r.status)
	require.NoError(t, svc.ProcessRoundTimeout(ctx, map[string]interface{}{"gameId": "g", "round": float64(1)}))
	require.Equal(t, 3, r.opens)
}
