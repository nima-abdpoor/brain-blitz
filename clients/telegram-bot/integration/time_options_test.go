//go:build integration

package integration

import (
	"context"
	apihttp "github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/http"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/auth"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestIntegration_TimeOptionsAndAuthorization(t *testing.T) {
	requireBackend(t, userServiceURL())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	sessions := session.NewMemoryStore()
	user := apihttp.NewUserClient(apihttp.NewClient(userServiceURL(), 5*time.Second, discardLogger()))
	svc := auth.NewService(user, sessions)
	if _, err := svc.Register(ctx, 9001, uniqueEmail("time-options"), integrationPassword); err != nil {
		t.Fatal(err)
	}
	sess, err := sessions.Get(ctx, 9001)
	if err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSuffix(userServiceURL(), "/user-service") + "/question-service"
	client := apihttp.NewTimeOptionsClient(apihttp.NewClient(base, 5*time.Second, discardLogger()))
	options, err := client.TimeOptions(ctx, sess.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]int{}
	for _, o := range options {
		values[o.ID] = o.Seconds
	}
	if values["10-seconds"] != 10 || values["1-minute"] != 60 {
		t.Fatalf("missing seeded timers: %+v", options)
	}
	for _, test := range []struct {
		path, method, token string
		want                int
	}{{"/api/v1/question/time-options", "GET", "", 401}, {"/api/v1/question/admin/time-options", "GET", sess.AccessToken, 403}, {"/api/v1/question/admin/time-options/10-seconds", "PUT", sess.AccessToken, 403}} {
		req, err := http.NewRequestWithContext(ctx, test.method, base+test.path, strings.NewReader(`{"label":"Bad edit","seconds":5,"enabled":true}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		if test.token != "" {
			req.Header.Set("Authorization", "Bearer "+test.token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != test.want {
			t.Errorf("%s %s status=%d want %d", test.method, test.path, resp.StatusCode, test.want)
		}
	}
}
