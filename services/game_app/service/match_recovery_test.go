package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/matchqueue"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type matchRecoveryRepository struct {
	roundRepository
	game       Game
	ready      uint64
	queueExits int
	cancelled  int
}

func (r *matchRecoveryRepository) LatestGameID(_ context.Context, playerID uint64) (string, error) {
	for _, member := range r.game.Players {
		if member == playerID {
			return "game", nil
		}
	}
	return "", nil
}

func (r *matchRecoveryRepository) GetGame(context.Context, string) (Game, error) {
	return r.game, nil
}

func (r *matchRecoveryRepository) IsPlayerReady(_ context.Context, _ string, id uint64) (bool, error) {
	return r.ready == id, nil
}

func (r *matchRecoveryRepository) CancelGame(context.Context, string) error {
	r.cancelled++
	r.game.Status = GameStatusCancelled
	return nil
}

func (r *matchRecoveryRepository) JoinQueue(context.Context, uint64, string, ...matchqueue.Choice) error {
	return nil
}
func (r *matchRecoveryRepository) LeaveQueue(context.Context, uint64) error {
	r.queueExits++
	return nil
}
func (r *matchRecoveryRepository) IsQueued(context.Context, uint64, string, ...matchqueue.Choice) bool {
	return false
}

func TestMatchSnapshotRestoresOnlyTheAuthenticatedPlayersProgress(t *testing.T) {
	deadline := time.Now().Add(time.Minute).UTC()
	repo := &matchRecoveryRepository{
		roundRepository: roundRepository{
			q: GameQuestions{Players: []uint64{1, 2}, Questions: []Question{
				{Id: "q1", Content: "Earlier", CorrectAnswer: "a1"},
				{Id: "q2", Content: "Current", CorrectAnswer: "secret-correct-choice", ValidAnswerTime: deadline},
			}},
			answers: map[string][]QuestionCorrectness{
				"1": {{QuestionId: "q1", PlayerChoice: "a1"}, {QuestionId: "q2", PlayerChoice: "my-answer"}},
				"2": {{QuestionId: "q2", PlayerChoice: "another-players-answer"}},
			},
		},
		game: Game{Players: []uint64{1, 2}, Status: GameStatusPending, ReadyUntil: deadline,
			PlayerNames: map[string]string{"1": "One", "2": "Two"}, Category: []Category{CategoryTypeTech}},
		ready: 1,
	}
	socket := &completionSocket{}
	svc := testService()
	svc.repository, svc.webSocket = repo, socket
	ctx := context.Background()

	require.NoError(t, svc.syncMatch(ctx, 9, nil))
	require.Nil(t, socket.messages[0].MetaData.Snapshot, "an outsider's snapshot has no match")
	require.NoError(t, svc.syncMatch(ctx, 1, nil))
	ready := socket.messages[1].MetaData.Snapshot
	require.Equal(t, "ready", ready.Phase)
	require.True(t, ready.Ready)
	require.Equal(t, deadline, *ready.ReadyUntil)

	repo.q.RoundBased, repo.q.CurrentQuestionIndex = true, 1
	require.NoError(t, svc.syncMatch(ctx, 1, nil))
	active := socket.messages[2].MetaData.Snapshot
	require.Equal(t, "round", active.Phase)
	require.Equal(t, 2, active.RoundNumber)
	require.Equal(t, deadline, active.Question.TTL)
	require.Equal(t, &RecoveredSubmission{Choice: "my-answer"}, active.Submission)
	require.Equal(t, "a1", active.Recap.CorrectAnswer)
	body, err := json.Marshal(active)
	require.NoError(t, err)
	require.NotContains(t, string(body), "secret-correct-choice")
	require.NotContains(t, string(body), "another-players-answer")

	repo.game.Status, repo.game.UpdatedAt = GameStatusFinished, time.Now()
	repo.q = GameQuestions{Players: []uint64{1, 2}}
	require.NoError(t, svc.syncMatch(ctx, 1, nil))
	finished := socket.messages[3].MetaData.Snapshot
	require.Equal(t, "finished", finished.Phase)
	require.Len(t, finished.Scores, 2)
	require.Equal(t, "One", finished.Scores[0].DisplayName)
	require.Nil(t, finished.Question)
	require.Nil(t, finished.Scores[0].Answers)
}

func TestSnapshotWaitsForARoundTransition(t *testing.T) {
	repo := &matchRecoveryRepository{
		game: Game{Players: []uint64{1}, Status: GameStatusPending, ReadyUntil: time.Now().Add(time.Minute)},
		roundRepository: roundRepository{q: GameQuestions{
			Players: []uint64{1}, RoundBased: true, Questions: []Question{{Id: "q1"}, {Id: "q2"}},
		}},
	}
	socket := &completionSocket{}
	svc := testService()
	svc.repository, svc.webSocket = repo, socket
	lock := svc.roundMutex("game")
	lock.Lock()
	done := make(chan error, 1)
	go func() { done <- svc.syncMatch(context.Background(), 1, nil) }()
	repo.q.CurrentQuestionIndex = 1
	lock.Unlock()
	require.NoError(t, <-done)
	require.Equal(t, "q2", socket.messages[0].MetaData.Snapshot.Question.Id)
}

func TestReadyExpiryAndQueueExitReportTheCommittedOutcome(t *testing.T) {
	repo := &matchRecoveryRepository{
		game: Game{Players: []uint64{1}, Status: GameStatusPending, ReadyUntil: time.Now().Add(time.Minute)},
	}
	socket := &completionSocket{}
	svc := testService()
	svc.repository, svc.webSocket, svc.taskPublisher = repo, socket, roundPublisher{}
	ctx := context.Background()
	conn, peer := net.Pipe()
	t.Cleanup(func() { _ = conn.Close(); _ = peer.Close() })
	svc.connections[1] = conn
	require.NoError(t, svc.acknowledgeQueueExit(ctx, 1, conn))
	require.Equal(t, EventMatchCreated, socket.messages[0].Event)

	payload := map[string]interface{}{"gameId": "game"}
	require.NoError(t, svc.ProcessReadyTimeout(ctx, payload))
	require.Zero(t, repo.cancelled)
	repo.game.ReadyUntil = time.Now().Add(-time.Minute)
	repo.q.RoundBased = true
	require.NoError(t, svc.ProcessReadyTimeout(ctx, payload))
	require.Zero(t, repo.cancelled)
	repo.q.RoundBased = false
	require.NoError(t, svc.ProcessReadyTimeout(ctx, payload))
	require.NoError(t, svc.ProcessReadyTimeout(ctx, payload))
	require.Equal(t, 1, repo.cancelled)
	require.Equal(t, EventMatchCancelled, socket.messages[1].Event)
	require.NoError(t, svc.acknowledgeQueueExit(ctx, 1, conn))
	require.Equal(t, EventLeftQueue, socket.messages[2].Event)
	require.Equal(t, 2, repo.queueExits)
}

func TestSupersededSocketCannotSubmitAnAnswer(t *testing.T) {
	svc := testService()
	svc.logger = logger.SlogAdapter{Logger: slog.Default()}
	server, peer := net.Pipe()
	t.Cleanup(func() { _ = server.Close(); _ = peer.Close() })
	svc.connections[1] = server
	var superseded net.Conn
	err := svc.readMessage(context.Background(), 1, &superseded, websocket.OpText,
		`{"command":"ANSWER","answer":{"gameId":"game","questionId":"q1","choice":"x"}}`)
	require.ErrorIs(t, err, net.ErrClosed)
}
