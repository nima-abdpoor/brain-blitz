package service

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"

	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/playercount"
	"github.com/stretchr/testify/require"
)

func TestPlayerCountOptionsFollowConfiguration(t *testing.T) {
	svc := testService()
	require.Equal(t, 2, svc.GameOptions().MinPlayers)
	require.Equal(t, 15, svc.GameOptions().MaxPlayers)
	require.Len(t, svc.getCategories().NumberOfPlayers, 14)
	svc = NewService(Config{Players: playercount.Limits{Min: 3, Max: 4}}, nil, nil, nil, nil, logger.SlogAdapter{})
	require.Equal(t, []int{3, 4}, svc.GameOptions().NumberOfPlayers)
	require.Equal(t, []int{3, 4}, svc.getCategories().NumberOfPlayers)
	require.Panics(t, func() {
		NewService(Config{Players: playercount.Limits{Min: 5, Max: 3}}, nil, nil, nil, nil, logger.SlogAdapter{})
	})
}

func TestUnsupportedPlayerCountRejectedBeforeQueueOrPublish(t *testing.T) {
	for _, request := range []string{
		`{"command":"ADD_TO_WAITING_LIST","category":"TECH","players":1}`,
		`{"command":"ADD_TO_WAITING_LIST","category":"TECH","players":16}`,
		`{"command":"ADD_TO_WAITING_LIST","category":"TECH","players":-3}`,
	} {
		socket := &completionSocket{}
		// Nil broker/repository intentionally fail if validation touches the queue.
		svc := NewService(Config{}, nil, socket, nil, nil, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		var conn net.Conn
		require.NoError(t, svc.readMessage(context.Background(), 1, &conn, websocket.OpText, request))
		require.Len(t, socket.messages, 1)
		require.Equal(t, Error, socket.messages[0].Event)
		require.False(t, socket.messages[0].Success)
		require.Contains(t, socket.messages[0].Message, "between 2 and 15")
	}
}

func TestReadyRejectsPlayerOutsideRoster(t *testing.T) {
	socket := &completionSocket{}
	repo := &roundRepository{q: GameQuestions{Players: []uint64{1, 2, 3}}}
	svc := NewService(Config{}, repo, socket, nil, nil, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	var conn net.Conn
	require.NoError(t, svc.readMessage(context.Background(), 99, &conn, websocket.OpText, `{"command":"READY","gameId":"g"}`))
	require.Len(t, socket.messages, 1)
	require.Equal(t, Error, socket.messages[0].Event)
	require.Contains(t, socket.messages[0].Message, "not in this game")
}

func TestMatchNotificationContinuesAfterDisconnectedPlayer(t *testing.T) {
	socket := &completionSocket{}
	svc := NewService(Config{}, nil, socket, nil, nil, logger.SlogAdapter{})
	for _, id := range []uint64{1, 3} {
		a, b := net.Pipe()
		defer a.Close()
		defer b.Close()
		svc.connections[id] = a
	}
	require.Error(t, svc.writeMessage([]uint64{1, 2, 3}, ProcessGameMessageResponse{Event: EventMatchCreated}))
	require.Len(t, socket.messages, 2, "one disconnect must not prevent later players from receiving their match")
}
