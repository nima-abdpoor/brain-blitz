package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestRecapOnlyAfterRoundClosureAndBeforeNextQuestion(t *testing.T) {
	repo := &roundRepository{q: GameQuestions{RoundBased: true, Players: []uint64{1, 2}, Questions: []Question{{Id: "q1", Content: "First?", CorrectAnswer: "right"}, {Id: "q2", Content: "Second?", CorrectAnswer: "second"}}}, answers: map[string][]QuestionCorrectness{}}
	sock := &completionSocket{}
	svc := NewService(Config{}, repo, sock, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	for _, id := range []uint64{1, 2} {
		a, b := net.Pipe()
		defer a.Close()
		defer b.Close()
		svc.connections[id] = a
	}
	send := func(id uint64, choice string) {
		data, _ := json.Marshal(ProcessGameMessageRequest{Command: CommandAnswer, GameAnswer: ProcessGameAnswer{GameId: "g", QuestionId: "q1", Answer: choice}})
		conn := svc.connections[id]
		require.NoError(t, svc.readMessage(context.Background(), id, &conn, websocket.OpText, string(data)))
	}
	send(1, "wrong")
	require.Len(t, sock.messages, 1)
	require.Equal(t, AnswerAccepted, sock.messages[0].Event)
	require.Equal(t, &ProcessGameAnswerFeedback{QuestionId: "q1", IsCorrect: false}, sock.messages[0].MetaData.AnswerFeedback)
	for _, p := range sock.messages[0].MetaData.Answer.PlayerPoint {
		require.Empty(t, p.Answers, "ack must not disclose current answers")
	}
	send(2, "right")
	require.Len(t, sock.messages, 6)
	require.Equal(t, AnswerAccepted, sock.messages[1].Event)
	require.Equal(t, &ProcessGameAnswerFeedback{QuestionId: "q1", IsCorrect: true}, sock.messages[1].MetaData.AnswerFeedback)
	correctRecaps := 0
	for _, message := range sock.messages[2:4] {
		require.Equal(t, EventRoundClosed, message.Event)
		require.Equal(t, "right", message.MetaData.Recap.CorrectAnswer)
		require.True(t, message.MetaData.Recap.Answered)
		if message.MetaData.Recap.IsCorrect {
			correctRecaps++
		}
	}
	require.Equal(t, 1, correctRecaps)
	require.Equal(t, NewQuestion, sock.messages[4].Event)
	require.Equal(t, NewQuestion, sock.messages[5].Event)
	repo.q.Questions[1].ValidAnswerTime = time.Now().Add(-time.Second)
	require.NoError(t, svc.ProcessRoundTimeout(context.Background(), map[string]interface{}{"gameId": "g", "round": float64(1)}))
	for _, message := range sock.messages[6:8] {
		require.Equal(t, EventRoundClosed, message.Event)
		require.False(t, message.MetaData.Recap.Answered)
		require.Equal(t, "second", message.MetaData.Recap.CorrectAnswer)
	}
	require.Equal(t, EventCompleted, sock.messages[8].Event)
}
