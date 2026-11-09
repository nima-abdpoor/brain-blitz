package matchmaking

import (
	"context"
	"fmt"
)

type PlayerCountsAPI interface {
	PlayerCounts(context.Context, string) ([]int, error)
}

func (m *Manager) SetPlayerCountsAPI(api PlayerCountsAPI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerCounts = api
}

func (m *Manager) PlayerCounts(ctx context.Context, chatID int64) ([]int, error) {
	sess, err := m.sessions.Get(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if sess.Stale() {
		return nil, ErrSessionExpired
	}
	m.mu.Lock()
	api := m.playerCounts
	m.mu.Unlock()
	if api == nil {
		return nil, fmt.Errorf("player counts unavailable")
	}
	return api.PlayerCounts(ctx, sess.AccessToken)
}

func (m *Manager) JoinQueueForPlayers(ctx context.Context, chatID int64, category, timeOptionID string, players int) error {
	counts, err := m.PlayerCounts(ctx, chatID)
	if err != nil {
		return err
	}
	for _, count := range counts {
		if count == players {
			return m.joinQueue(ctx, chatID, category, timeOptionID, players)
		}
	}
	return fmt.Errorf("player count is no longer available; choose again")
}
