//go:build integration

package integration

import (
	"context"
	apihttp "github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/apiclient/http"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/auth"
	"github.com/nima-abdpoor/brain-blitz-telegram-bot/internal/core/session"
	"testing"
	"time"
)

func TestIntegration_UnknownAccountIsInvalidCredentials(t *testing.T) {
	requireBackend(t, userServiceURL())
	client := apihttp.NewUserClient(apihttp.NewClient(userServiceURL(), 5*time.Second, discardLogger()))
	svc := auth.NewService(client, session.NewMemoryStore())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := svc.Login(ctx, 1, uniqueEmail("nonexistent"), integrationPassword)
	authErr, ok := err.(*auth.Error)
	if !ok || authErr.Kind != auth.KindInvalidCredentials {
		t.Fatalf("unknown account = %v, want invalid credentials", err)
	}
}
