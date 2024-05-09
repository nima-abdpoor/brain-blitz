package presence

import (
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"time"
)

type Presence struct {
	db *redis.Adapter
}

func New(db *redis.Adapter) repository.PresenceRepository {
	return Presence{
		db: db,
	}
}

func (p Presence) Upsert(ctx context.Context, key string, timestamp int64, expTime time.Duration) error {
	const op = "presence.Upsert"
	if _, err := p.db.Client().Set(ctx, key, timestamp, expTime).Result(); err != nil {
		fmt.Println(op, err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	}
	return nil
}
