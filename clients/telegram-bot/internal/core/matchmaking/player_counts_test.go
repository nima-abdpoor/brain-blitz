package matchmaking

import (
	"context"
	"fmt"
	"testing"
)

type playerCountsAPI struct {
	counts []int
	err    error
}

func (a playerCountsAPI) PlayerCounts(context.Context, string) ([]int, error) { return a.counts, a.err }

func TestSelectedPlayerCountSentToBackend(t *testing.T) {
	for _, n := range []int{2, 3, 4, 15} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			conn := newFakeConn()
			defer conn.endListen()
			mgr, sessions := newTestManager(t, conn, nil, newFakeNotifier())
			mustSaveSession(t, sessions, testChatID)
			mgr.SetPlayerCountsAPI(playerCountsAPI{counts: []int{2, 3, 4, 15}})
			if err := mgr.JoinQueueForPlayers(context.Background(), testChatID, "TECH", "fast", n); err != nil {
				t.Fatal(err)
			}
			defer mgr.Detach(testChatID)
			commands := conn.sentCommands()
			if len(commands) != 1 || commands[0].Players != n || commands[0].TimeOptionID != "fast" || commands[0].Category != "TECH" {
				t.Fatalf("wrong selection sent: %+v", commands)
			}
		})
	}
}

func TestUnavailablePlayerCountsDoNotJoin(t *testing.T) {
	for _, api := range []playerCountsAPI{{counts: []int{3, 4}}, {err: fmt.Errorf("unavailable")}} {
		conn := newFakeConn()
		mgr, sessions := newTestManager(t, conn, nil, newFakeNotifier())
		mustSaveSession(t, sessions, testChatID)
		mgr.SetPlayerCountsAPI(api)
		if err := mgr.JoinQueueForPlayers(context.Background(), testChatID, "TECH", "fast", 15); err == nil {
			t.Fatal("expected rejection")
		}
		if mgr.StateFor(testChatID) != Idle || len(conn.sentCommands()) != 0 {
			t.Fatal("invalid count must not open a queue")
		}
	}
}
