package matchmaking

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"fmt"
	"time"
)

type MatchMaking struct {
	db     *redis.Adapter
	config Config
}

type Config struct {
	WaitingListPrefix string `koanf:"waitingListPrefix"`
}

func NewMatchMakingRepo(db *redis.Adapter, config Config) repository.MatchMakingRepository {
	return MatchMaking{
		db:     db,
		config: config,
	}
}

func (m MatchMaking) AddToWaitingList(category entity.Category, userId int64) error {
	fmt.Println("here1:", fmt.Sprintf("%s:%v", m.config.WaitingListPrefix, category))
	err := redis.ZAdd(m.db.Client(),
		fmt.Sprintf("%s:%v", m.config.WaitingListPrefix, category),
		float64(time.Now().UnixMicro()),
		userId,
	)
	return err
}
