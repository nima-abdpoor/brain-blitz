package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type completionRepository struct {
	Repository
	finished bool
	statuses map[uint64]GameStatus
}

func (r *completionRepository) GetGame(context.Context, string) (Game, error) {
	return Game{Players: []uint64{1, 2, 3}}, nil
}
func (r *completionRepository) GetLeaderBoard(context.Context, string) (LeaderBoard, error) {
	return LeaderBoard{PlayersPoint: []PlayerPoint{{PlayerId: "1", Point: 5}}}, nil
}
func (r *completionRepository) FinishGame(context.Context, string) error {
	r.finished = true
	return nil
}
func (r *completionRepository) UpsertUserStatus(_ context.Context, id uint64, status GameStatus) error {
	r.statuses[id] = status
	return nil
}

type completionSocket struct {
	websocket.WebSocket
	messages []ProcessGameMessageResponse
}

func (s *completionSocket) WriteServerData(_ io.Writer, _ websocket.OpCode, value string) error {
	var msg ProcessGameMessageResponse
	if err := json.Unmarshal([]byte(value), &msg); err != nil {
		return err
	}
	s.messages = append(s.messages, msg)
	return nil
}

func TestCompletionIncludesUnansweredAndDisconnectedPlayers(t *testing.T) {
	repo := &completionRepository{statuses: make(map[uint64]GameStatus)}
	socket := &completionSocket{}
	svc := NewService(Config{}, repo, socket, nil, nil, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	for _, id := range []uint64{1, 2} {
		server, client := net.Pipe()
		defer server.Close()
		defer client.Close()
		svc.connections[id] = server
	}
	require.NoError(t, svc.ProcessGameCompletion(context.Background(), map[string]interface{}{"gameId": "game-1"}))
	assert.True(t, repo.finished)
	assert.Equal(t, map[uint64]GameStatus{1: GameStatusFinished, 2: GameStatusFinished, 3: GameStatusFinished}, repo.statuses)
	require.Len(t, socket.messages, 2)
	for _, msg := range socket.messages {
		assert.Equal(t, EventCompleted, msg.Event)
		assert.True(t, msg.Success)
	}
	assert.Empty(t, svc.connections)
}

func TestOldConnectionCleanupKeepsReplacement(t *testing.T) {
	svc := testService()
	old, oldClient := net.Pipe()
	replacement, newClient := net.Pipe()
	defer oldClient.Close()
	defer replacement.Close()
	defer newClient.Close()
	svc.connections[1] = replacement
	svc.closeConnection(1, old)
	assert.Equal(t, replacement, svc.connections[1])
}
