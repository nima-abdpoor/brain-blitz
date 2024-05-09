package repository

import (
	"context"
	"time"
)

type PresenceRepository interface {
	Upsert(ctx context.Context, key string, timestamp int64, expTime time.Duration) error
}
