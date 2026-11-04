package repository

import (
	"BrainBlitz.com/game/pkg/matchqueue"
	"context"
	"strconv"
)

func (m GameRepository) JoinQueue(ctx context.Context, id uint64, category string, choices ...matchqueue.Choice) error {
	return matchqueue.Join(ctx, m.redisDB.Client(), strconv.FormatUint(id, 10), category, choices...)
}
func (m GameRepository) LeaveQueue(ctx context.Context, id uint64) error {
	return matchqueue.Leave(ctx, m.redisDB.Client(), "waiting_users", strconv.FormatUint(id, 10))
}
func (m GameRepository) IsQueued(ctx context.Context, id uint64, category string, choices ...matchqueue.Choice) bool {
	return matchqueue.Active(ctx, m.redisDB.Client(), strconv.FormatUint(id, 10), category, choices...)
}
