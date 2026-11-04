package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/matchqueue"
	"BrainBlitz.com/game/services/match_app/service"
	"context"
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	WaitingListPrefix           string        `koanf:"waitingListPrefix"`
	MinTimeWaitingListSelection time.Duration `koanf:"min_time_list_selection"`
}

type MatchMakingRepository struct {
	Config Config
	Logger logger.SlogAdapter
	db     *redis.Adapter
}

func NewRepository(config Config, logger logger.SlogAdapter, redis *redis.Adapter) service.Repository {
	return MatchMakingRepository{
		Config: config,
		Logger: logger,
		db:     redis,
	}
}

func (m MatchMakingRepository) AddToWaitingList(ctx context.Context, category service.Category, userId string) error {
	return matchqueue.Enqueue(ctx, m.db.Client(), m.Config.WaitingListPrefix, userId, string(category))
}

func (m MatchMakingRepository) GetWaitingListByCategory(ctx context.Context, category service.Category) ([]service.WaitingMember, error) {
	key := fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, category)
	mintTime := int(time.Now().Add(m.Config.MinTimeWaitingListSelection).UnixMicro())
	maxTime := int(time.Now().UnixMicro())
	if err := m.db.Client().ZRemRangeByScore(ctx, key, "-inf", strconv.Itoa(mintTime)).Err(); err != nil {
		return nil, err
	}
	if err, list := m.db.ZRange(ctx, key, mintTime, maxTime, true); err != nil {
		return nil, err
	} else {
		result := make([]service.WaitingMember, 0)
		for _, z := range list {
			if !matchqueue.Active(ctx, m.db.Client(), z.Member.(string), string(category)) {
				continue
			}
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

func (m MatchMakingRepository) RemoveWaitingMember(ctx context.Context, members []service.WaitingMember) error {
	for _, member := range members {
		key := fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, member.Category)
		err := m.db.ZDelete(ctx, key, redis.Z{
			Member: member.UserId,
		})
		if err != nil {
			m.Logger.Error("error in deleting user from waiting list", "key", key, "err", err.Error())
		}
	}

	return nil
}
