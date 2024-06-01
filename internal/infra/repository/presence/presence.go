package presence

import (
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"strconv"
	"time"
)

type Presence struct {
	db     *redis.Adapter
	config Config
}

type Config struct {
	PresencePrefix string `koanf:"presence_prefix"`
}

func New(db *redis.Adapter, config Config) repository.PresenceRepository {
	return Presence{
		db:     db,
		config: config,
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

func (p Presence) GetPresenceByUserID(context context.Context, userId string) (int64, error) {
	const op = "presence.GetPresenceByUserID"
	if res, err := p.db.Client().Get(context, fmt.Sprintf("%s:%s", p.config.PresencePrefix, userId)).Result(); err != nil {
		fmt.Println(op, err)
		return 0, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	} else {
		timeStamp, _ := strconv.Atoi(res)
		return int64(timeStamp), nil
	}
}

func (p Presence) GetPresence(context context.Context, userIds []string) (map[string]int64, error) {
	const op = "presence.GetPresence"

	keys := make([]string, len(userIds))
	for i, userId := range userIds {
		keys[i] = fmt.Sprintf("%s:%s", p.config.PresencePrefix, userId)
	}
	if res, err := p.db.Client().MGet(context, keys...).Result(); err != nil {
		fmt.Println(op, err)
		return map[string]int64{}, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	} else {
		results := make(map[string]int64, len(userIds))
		for i, val := range res {
			if val != nil {
				timeStamp, _ := strconv.Atoi(val.(string))
				results[userIds[i]] = int64(timeStamp)
			} else {
				results[userIds[i]] = 0
			}
		}
		return results, nil
	}
}
