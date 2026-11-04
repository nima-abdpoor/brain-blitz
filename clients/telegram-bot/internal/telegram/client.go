package telegram

import (
	"net/http"
	"strings"
	"time"
)

type safeHTTPClient struct {
	token                      string
	regular, polling, callback *http.Client
}

type safeRequestError struct {
	err   error
	token string
}

func (e safeRequestError) Error() string {
	return strings.ReplaceAll(e.err.Error(), e.token, "[REDACTED_BOT_TOKEN]")
}
func (e safeRequestError) Unwrap() error { return e.err }

// Polls need more than Telegram's 30-second long-poll window; interactive
// requests are bounded independently so an outage cannot freeze every chat.
func NewHTTPClient(token string) *safeHTTPClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	return &safeHTTPClient{token: token,
		regular:  &http.Client{Transport: transport, Timeout: 8 * time.Second},
		polling:  &http.Client{Transport: transport, Timeout: 40 * time.Second},
		callback: &http.Client{Transport: transport, Timeout: 2 * time.Second}}
}

func (c *safeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	client := c.regular
	if strings.HasSuffix(req.URL.Path, "/getUpdates") {
		client = c.polling
	}
	if strings.HasSuffix(req.URL.Path, "/answerCallbackQuery") {
		client = c.callback
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, safeRequestError{err, c.token}
	}
	return resp, nil
}
