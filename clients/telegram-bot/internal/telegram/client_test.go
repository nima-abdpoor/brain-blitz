package telegram

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNetworkErrorsRedactTokenAndPreserveCause(t *testing.T) {
	const token = "12345:fake-secret-token"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1/bot"+token+"/getUpdates", nil)
	_, err := NewHTTPClient(token).Do(req)
	if err == nil || strings.Contains(err.Error(), token) || !strings.Contains(err.Error(), "REDACTED") || !errors.Is(err, context.Canceled) {
		t.Fatalf("unsafe or unexpected error: %v", err)
	}
}
