package matchmaking

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"context"
	"fmt"
	"strconv"
	"time"
)

type MatchMaking struct {
	db     *redis.Adapter
	config Config
}

type Config struct {
	WaitingListPrefix           string        `koanf:"waitingListPrefix"`
	MinTimeWaitingListSelection time.Duration `koanf:"mint_time_list_selection"`
}

func NewMatchMakingRepo(db *redis.Adapter, config Config) repository.MatchMakingRepository {
	return MatchMaking{
		db:     db,
		config: config,
	}
}

func (m MatchMaking) AddToWaitingList(category entity.Category, userId string) error {
	err := redis.ZAdd(m.db.Client(),
		fmt.Sprintf("%s:%v", m.config.WaitingListPrefix, category),
		float64(time.Now().UnixMicro()),
		userId,
	)
	return err
}

func (m MatchMaking) GetWaitingListByCategory(ctx context.Context, category entity.Category) ([]entity.WaitingMember, error) {
	key := fmt.Sprintf("%s:%v", m.config.WaitingListPrefix, category)
	mintTime := strconv.Itoa(int(time.Now().Add(m.config.MinTimeWaitingListSelection).UnixMicro()))
	maxTime := strconv.Itoa(int(time.Now().UnixMicro()))
	if list, err := redis.ZRange(ctx, m.db.Client(), key, mintTime, maxTime); err != nil {
		return nil, err
	} else {
		result := make([]entity.WaitingMember, 0)
		for _, z := range list {
			userId, _ := strconv.Atoi(z.Member.(string))
			result = append(result, entity.WaitingMember{
				UserId:    uint(userId),
				TimeStamp: int64(z.Score),
				Category:  category,
			})
		}
		return result, nil
	}
}
