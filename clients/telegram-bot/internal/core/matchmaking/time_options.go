package matchmaking

import (
	"context"
	"fmt"
)

type TimeOption struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Seconds int    `json:"seconds"`
	Enabled bool   `json:"enabled"`
}
type TimeOptionsAPI interface {
	TimeOptions(context.Context, string) ([]TimeOption, error)
}

func (m *Manager) SetTimeOptionsAPI(api TimeOptionsAPI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeOptions = api
}
func (m *Manager) TimeOptions(ctx context.Context, chatID int64) ([]TimeOption, error) {
	sess, err := m.sessions.Get(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if sess.Stale() {
		return nil, ErrSessionExpired
	}
	m.mu.Lock()
	api := m.timeOptions
	m.mu.Unlock()
	if api == nil {
		return nil, fmt.Errorf("question time options unavailable")
	}
	return api.TimeOptions(ctx, sess.AccessToken)
}
