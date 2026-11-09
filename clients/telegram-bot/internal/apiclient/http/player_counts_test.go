package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPlayerCountsFetchesAuthenticatedBackendOptions(t *testing.T) {
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`{"minPlayers":3,"maxPlayers":4,"numberOfPlayers":[3,4]}`, true},
		{`{"minPlayers":2,"maxPlayers":4,"numberOfPlayers":[2,4]}`, false},
		{`{"minPlayers":1,"maxPlayers":1,"numberOfPlayers":[1]}`, false},
		{`{}`, false},
		{`broken`, false},
	} {
		t.Run(tc.body, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/game-options" || r.Header.Get("Authorization") != "Bearer token" {
					t.Errorf("invalid request: %s", r.URL)
				}
				fmt.Fprint(w, tc.body)
			}))
			defer s.Close()
			counts, err := NewPlayerCountsClient(NewClient(s.URL, time.Second, testLogger(), WithMaxRetries(0))).PlayerCounts(context.Background(), "token")
			if (err == nil) != tc.valid {
				t.Fatalf("counts=%v error=%v", counts, err)
			}
			if tc.valid && (len(counts) != 2 || counts[0] != 3 || counts[1] != 4) {
				t.Fatal(counts)
			}
		})
	}
}
