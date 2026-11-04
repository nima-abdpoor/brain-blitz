package game

import (
	"encoding/json"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/ws"
	"io"
	"log/slog"
	"testing"
)

type recapNotifier struct {
	*fakeNotifier
	count int
}

func (n *recapNotifier) NotifyRoundClosed(_ int64, _ RoundRecap, _ int) { n.count++ }
func TestRoundRecapIgnoresDuplicateAndWrongGame(t *testing.T) {
	n := &recapNotifier{fakeNotifier: newFakeNotifier()}
	m := NewManager(n, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	cg := &chatGame{gameID: "g", roundNumber: 1}
	m.chats[1] = cg
	send := func(id string, round int) {
		data, _ := json.Marshal(map[string]interface{}{"gameId": id, "roundNumber": round, "recap": RoundRecap{QuestionID: "q", CorrectAnswer: "right"}})
		m.handleRoundClosed(1, cg, ws.Event{MetaData: data})
	}
	send("other", 1)
	send("g", 2)
	send("g", 1)
	send("g", 1)
	if n.count != 1 {
		t.Fatalf("recap count=%d want 1", n.count)
	}
}
