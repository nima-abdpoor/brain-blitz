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
	"strconv"
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
	r.answers[a.PlayerID] = append(r.answers[a.PlayerID], QuestionCorrectness{QuestionId: a.QuestionIDs, PlayerChoice: a.PlayerChoice})
	return nil
}
func (r *roundRepository) GetLeaderBoard(context.Context, string) (LeaderBoard, error) {
	b := LeaderBoard{}
	for _, playerID := range r.q.Players {
		id := strconv.FormatUint(playerID, 10)
		b.PlayersPoint = append(b.PlayersPoint, PlayerPoint{PlayerId: id, QuestionCorrectness: r.answers[id]})
	}
	return b, nil
}
func (r *roundRepository) GetGame(context.Context, string) (Game, error) {
	return Game{Players: r.q.Players, Status: r.status}, nil
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
	for _, count := range []int{2, 3, 4, 15} {
		t.Run(fmt.Sprint(count), func(t *testing.T) { testConcurrentAnswers(t, count) })
	}
}
func testConcurrentAnswers(t *testing.T, count int) {
	players := make([]uint64, count)
	for i := range players {
		players[i] = uint64(i + 1)
	}
	r := &roundRepository{q: GameQuestions{Players: players, Questions: []Question{{Id: "q1"}, {Id: "q2"}}}, answers: map[string][]QuestionCorrectness{}}
	sock := &completionSocket{}
	svc := NewService(Config{}, r, sock, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	ctx := context.Background()
	require.NoError(t, svc.startRounds(ctx, "g"))
	deadline := r.q.Questions[0].ValidAnswerTime
	require.NoError(t, svc.startRounds(ctx, "g"))
	require.Equal(t, 1, r.opens)
	require.Equal(t, deadline, r.q.Questions[0].ValidAnswerTime)
	var wg sync.WaitGroup
	for _, id := range append(append([]uint64{}, players...), players...) {
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

func TestRoundWaitsForEntireRoster(t *testing.T) {
	for _, count := range []int{3, 4, 15} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			players := make([]uint64, count)
			for i := range players {
				players[i] = uint64(i + 1)
			}
			repo := &roundRepository{q: GameQuestions{Players: players, Questions: []Question{{Id: "q1"}, {Id: "q2"}}}, answers: map[string][]QuestionCorrectness{}}
			svc := NewService(Config{}, repo, &completionSocket{}, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
			require.NoError(t, svc.startRounds(context.Background(), "g"))
			for i, id := range players {
				var conn net.Conn
				data, err := json.Marshal(ProcessGameMessageRequest{Command: CommandAnswer, GameAnswer: ProcessGameAnswer{GameId: "g", QuestionId: "q1", Answer: "wrong"}})
				require.NoError(t, err)
				require.NoError(t, svc.readMessage(context.Background(), id, &conn, websocket.OpText, string(data)))
				if i < count-1 {
					require.Zero(t, repo.q.CurrentQuestionIndex, "must wait for every player")
				}
			}
			require.Equal(t, 1, repo.q.CurrentQuestionIndex)
		})
	}
}
