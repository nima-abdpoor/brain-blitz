package repository

import (
	"context"
	"time"
)

type PresenceRepository interface {
	Upsert(ctx context.Context, key string, timestamp int64, expTime time.Duration) error
}

type PresenceClient interface {
	GetPresenceByUserID(context context.Context, userId string) (int64, error)
	GetPresence(context context.Context, userId []string) (map[string]int64, error)
}
