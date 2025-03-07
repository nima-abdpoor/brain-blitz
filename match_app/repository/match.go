package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/match_app/service"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

type Config struct {
	WaitingListPrefix           string        `koanf:"waitingListPrefix"`
	MinTimeWaitingListSelection time.Duration `koanf:"min_time_list_selection"`
}

type MatchMakingRepository struct {
	Config Config
	Logger *slog.Logger
	db     *redis.Adapter
}

func NewRepository(config Config, logger *slog.Logger, redis *redis.Adapter) service.Repository {
	return MatchMakingRepository{
		Config: config,
		Logger: logger,
		db:     redis,
	}
}

func (m MatchMakingRepository) AddToWaitingList(ctx context.Context, category service.Category, userId string) error {
	return m.db.ZAdd(
		ctx,
		fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, category), redis.Z{
			Score:  float64(time.Now().UnixMicro()),
			Member: userId,
		})
}

func (m MatchMakingRepository) GetWaitingListByCategory(ctx context.Context, category service.Category) ([]service.WaitingMember, error) {
	key := fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, category)
	mintTime := int(time.Now().Add(m.Config.MinTimeWaitingListSelection).UnixMicro())
	maxTime := int(time.Now().UnixMicro())
	if err, list := m.db.ZRange(ctx, key, mintTime, maxTime, true); err != nil {
		return nil, err
	} else {
		result := make([]service.WaitingMember, 0)
		for _, z := range list {
			userId, _ := strconv.Atoi(z.Member.(string))
			result = append(result, service.WaitingMember{
				UserId:    uint(userId),
				TimeStamp: int64(z.Score),
				Category:  category,
			})
		}
		return result, nil
	}
}
