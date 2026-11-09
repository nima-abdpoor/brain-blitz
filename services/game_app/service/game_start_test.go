package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type startingRepository struct {
	Repository
	reads    int
	readyAt  int
	prepared bool
	startErr error
}

func (r *startingRepository) GetGame(context.Context, string) (Game, error) {
	return Game{Players: []uint64{1, 2}}, nil
}

func (r *startingRepository) GetQuestionsByGameId(context.Context, string) (GameQuestions, error) {
	r.reads++
	if r.reads < r.readyAt {
		return GameQuestions{}, errors.New("questions not delivered yet")
	}
	return GameQuestions{Questions: []Question{{Id: "q1"}}}, nil
}

func (r *startingRepository) SetValidAnswerTimeForQuestions(context.Context, string) (time.Duration, error) {
	r.prepared = true
	return 20 * time.Second, r.startErr
}

func (r *startingRepository) UpsertReadyPlayer(context.Context, string, *int, *int) (bool, error) {
	return true, nil
}

func TestPrepareGameWaitsForAsynchronousQuestions(t *testing.T) {
	repo := &startingRepository{readyAt: 2}
	svc := testService()
	svc.repository = repo
	ttl, err := svc.prepareGameQuestions(context.Background(), "game-1")
	require.NoError(t, err)
	require.Equal(t, 20*time.Second, ttl)
	require.Equal(t, 2, repo.reads)
	require.True(t, repo.prepared)
}

func TestPrepareGameRespectsCancellation(t *testing.T) {
	repo := &startingRepository{readyAt: 100}
	svc := testService()
	svc.repository = repo
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := svc.prepareGameQuestions(ctx, "game-1")
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, repo.prepared)
}

func TestReadyFailureDoesNotPublishImmediateCompletion(t *testing.T) {
	repo := &startingRepository{readyAt: 1, startErr: errors.New("could not set deadlines")}
	socket := &completionSocket{}
	log := logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	// A nil publisher deliberately fails the test if the error path falls
	// through and attempts to schedule game completion.
	svc := NewService(Config{}, repo, socket, nil, nil, log)
	var conn net.Conn
	require.NoError(t, svc.readMessage(context.Background(), 1, &conn, websocket.OpText,
		`{"command":"READY","gameId":"game-1"}`))
	require.Len(t, socket.messages, 1)
	require.Equal(t, Error, socket.messages[0].Event)
	require.False(t, socket.messages[0].Success)
}
