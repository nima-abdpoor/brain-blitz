package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type feedbackSocket struct {
	completionSocket
	recipients []io.Writer
}

func (s *feedbackSocket) WriteServerData(w io.Writer, code websocket.OpCode, value string) error {
	s.recipients = append(s.recipients, w)
	return s.completionSocket.WriteServerData(w, code, value)
}

func TestAnswerFeedbackArrivesBeforeOtherPlayersAnswer(t *testing.T) {
	for _, count := range []int{2, 3, 4, 15} {
		for _, choice := range []string{"right", "wrong"} {
			t.Run(fmt.Sprintf("%d_players/%s", count, choice), func(t *testing.T) {
				players := make([]uint64, count)
				for i := range players {
					players[i] = uint64(i + 1)
				}
				repo := &roundRepository{q: GameQuestions{RoundBased: true, Players: players, Questions: []Question{{Id: "q1", CorrectAnswer: "right", ValidAnswerTime: time.Now().Add(time.Hour)}}}, answers: map[string][]QuestionCorrectness{}}
				sock := &feedbackSocket{}
				svc := NewService(Config{}, repo, sock, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
				for _, id := range players {
					server, client := net.Pipe()
					t.Cleanup(func() { server.Close(); client.Close() })
					svc.connections[id] = server
				}
				conn := svc.connections[1]
				data, err := json.Marshal(ProcessGameMessageRequest{Command: CommandAnswer, GameAnswer: ProcessGameAnswer{GameId: "g", QuestionId: "q1", Answer: choice}})
				require.NoError(t, err)
				require.NoError(t, svc.readMessage(context.Background(), 1, &conn, websocket.OpText, string(data)))

				require.Len(t, sock.messages, 1, "feedback must arrive while the other players are still answering")
				require.Equal(t, []io.Writer{conn}, sock.recipients, "only the submitting player receives feedback")
				ack := sock.messages[0]
				require.Equal(t, AnswerAccepted, ack.Event)
				require.True(t, ack.Success)
				require.Equal(t, &ProcessGameAnswerFeedback{QuestionId: "q1", IsCorrect: choice == "right"}, ack.MetaData.AnswerFeedback)
				require.Zero(t, repo.opens, "feedback must not advance the round")
				require.Nil(t, ack.MetaData.Recap)
				for _, p := range ack.MetaData.Answer.PlayerPoint {
					require.Empty(t, p.Answers)
					// The test repository leaves all points at zero: the verdict
					// must compare choices rather than infer correctness from points.
					require.Zero(t, p.Point)
				}
				wire, err := json.Marshal(ack)
				require.NoError(t, err)
				require.NotContains(t, string(wire), "correctAnswer", "solutions stay hidden until round closure")
				require.Contains(t, string(wire), fmt.Sprintf(`"isCorrect":%t`, choice == "right"), "false verdicts must be present on the wire")

				// A rejected duplicate must not produce another success verdict.
				require.NoError(t, svc.readMessage(context.Background(), 1, &conn, websocket.OpText, string(data)))
				require.Len(t, sock.messages, 2)
				require.Equal(t, Error, sock.messages[1].Event)
				require.Nil(t, sock.messages[1].MetaData.AnswerFeedback)
			})
		}
	}
}

func TestRejectedAnswerDoesNotReceiveFeedback(t *testing.T) {
	for _, tc := range []struct {
		name       string
		playerID   uint64
		questionID string
	}{
		{name: "inactive question", playerID: 1, questionID: "old"},
		{name: "nonparticipant", playerID: 3, questionID: "q1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &roundRepository{q: GameQuestions{RoundBased: true, Players: []uint64{1, 2}, Questions: []Question{{Id: "q1", CorrectAnswer: "right"}}}, answers: map[string][]QuestionCorrectness{}}
			sock := &completionSocket{}
			svc := NewService(Config{}, repo, sock, nil, roundPublisher{}, logger.SlogAdapter{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
			var conn net.Conn
			data, err := json.Marshal(ProcessGameMessageRequest{Command: CommandAnswer, GameAnswer: ProcessGameAnswer{GameId: "g", QuestionId: tc.questionID, Answer: "right"}})
			require.NoError(t, err)
			require.NoError(t, svc.readMessage(context.Background(), tc.playerID, &conn, websocket.OpText, string(data)))
			require.Len(t, sock.messages, 1)
			require.Equal(t, Error, sock.messages[0].Event)
			require.Nil(t, sock.messages[0].MetaData.AnswerFeedback)
			require.Empty(t, repo.answers)
		})
	}
}
