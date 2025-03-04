package repository

import (
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/match_app/service"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

type Config struct {
	WaitingListPrefix           string        `koanf:"waitingListPrefix"`
	MinTimeWaitingListSelection time.Duration `koanf:"mint_time_list_selection"`
}

type MatchMakingRepository struct {
	Config Config
	Logger *slog.Logger
	db     *redis.Adapter
}

func NewUserRepository(config Config, logger *slog.Logger) service.Repository {
	return MatchMakingRepository{
		Config: config,
		Logger: logger,
	}
}

func (m MatchMakingRepository) AddToWaitingList(ctx context.Context, category service.Category, userId string) error {
	err := redis.ZAdd(m.db.Client(),
		fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, category),
		float64(time.Now().UnixMicro()),
		userId,
	)
	return err
}

func (m MatchMakingRepository) GetWaitingListByCategory(ctx context.Context, category service.Category) ([]service.WaitingMember, error) {
	key := fmt.Sprintf("%s:%v", m.Config.WaitingListPrefix, category)
	mintTime := strconv.Itoa(int(time.Now().Add(m.Config.MinTimeWaitingListSelection).UnixMicro()))
	maxTime := strconv.Itoa(int(time.Now().UnixMicro()))
	if list, err := redis.ZRange(ctx, m.db.Client(), key, mintTime, maxTime); err != nil {
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
