package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRepo wires up a GameRepository backed by miniredis (in-memory Redis).
func newTestRepo(t *testing.T) GameRepository {
	t.Helper()
	mr := miniredis.RunT(t)

	// redis.Adapter is a concrete struct; its client field is unexported, so we
	// use the exported constructor with the miniredis address.
	adapter := redis.New(redis.Config{
		Host:     mr.Host(),
		Port:     mr.Server().Addr().Port,
		Password: "",
		DB:       0,
	})

	return GameRepository{
		Config: Config{
			QuestionsTimeOut:   5 * time.Second,
			GameStatusTimeOut:  5 * time.Second,
			ValidAnswerTimeOut: 2 * time.Minute,
			ScoreConfig: ScoreConfig{
				BaseScore:     5,
				MaxBonus:      10,
				BonusDeadline: 115 * time.Second,
			},
		},
		Logger:  logger.SlogAdapter{},
		MongoDB: &mongo.Database{},
		redisDB: adapter,
	}
}

// pingRedis verifies the adapter can reach miniredis before running sub-tests.
func pingRedis(t *testing.T, adapter *redis.Adapter) {
	t.Helper()
	err := adapter.Client().Ping(context.Background()).Err()
	require.NoError(t, err, "miniredis must be reachable")
}

func TestUpsertReadyPlayer_TwoPlayerGame(t *testing.T) {
	repo := newTestRepo(t)
	pingRedis(t, repo.redisDB)

	ctx := context.Background()
	gameId := "game-001"
	expectedPlayers := 2
	p1, p2 := 101, 102

	t.Run("initial call with numberOfPlayers creates game status and returns false", func(t *testing.T) {
		ready, err := repo.UpsertReadyPlayer(ctx, gameId, nil, &expectedPlayers)
		require.NoError(t, err)
		assert.False(t, ready, "game must not be ready before any player sends READY")
	})

	t.Run("first player READY returns false", func(t *testing.T) {
		ready, err := repo.UpsertReadyPlayer(ctx, gameId, &p1, nil)
		require.NoError(t, err)
		assert.False(t, ready, "game must not be ready after only the first player sends READY")
	})

	t.Run("second player READY returns true", func(t *testing.T) {
		ready, err := repo.UpsertReadyPlayer(ctx, gameId, &p2, nil)
		require.NoError(t, err)
		assert.True(t, ready, "game must be ready after all players send READY")
	})

	t.Run("subsequent call returns true (idempotent full-roster guard)", func(t *testing.T) {
		ready, err := repo.UpsertReadyPlayer(ctx, gameId, &p1, nil)
		require.NoError(t, err)
		assert.True(t, ready, "a call when the roster is already full must still return true")
	})
}

func TestUpsertReadyPlayer_DuplicatePlayerRejected(t *testing.T) {
	repo := newTestRepo(t)
	pingRedis(t, repo.redisDB)

	ctx := context.Background()
	gameId := "game-002"
	expectedPlayers := 2
	p1 := 201

	_, err := repo.UpsertReadyPlayer(ctx, gameId, nil, &expectedPlayers)
	require.NoError(t, err)

	_, err = repo.UpsertReadyPlayer(ctx, gameId, &p1, nil)
	require.NoError(t, err)

	_, err = repo.UpsertReadyPlayer(ctx, gameId, &p1, nil)
	assert.Error(t, err, "sending READY twice for the same player must return an error")
	assert.Equal(t, fmt.Sprintf("player %v is already member of ready players", &p1), err.Error())
}

func TestUpsertReadyPlayer_ThreePlayerGame(t *testing.T) {
	repo := newTestRepo(t)
	pingRedis(t, repo.redisDB)

	ctx := context.Background()
	gameId := "game-003"
	expectedPlayers := 3
	p1, p2, p3 := 301, 302, 303

	_, err := repo.UpsertReadyPlayer(ctx, gameId, nil, &expectedPlayers)
	require.NoError(t, err)

	ready, err := repo.UpsertReadyPlayer(ctx, gameId, &p1, nil)
	require.NoError(t, err)
	assert.False(t, ready)

	ready, err = repo.UpsertReadyPlayer(ctx, gameId, &p2, nil)
	require.NoError(t, err)
	assert.False(t, ready, "game must not be ready after 2 of 3 players are ready")

	ready, err = repo.UpsertReadyPlayer(ctx, gameId, &p3, nil)
	require.NoError(t, err)
	assert.True(t, ready, "game must be ready after all 3 players send READY")
}

// Verify the redis.Adapter Port field access works with miniredis.
// miniredis.Server().Addr().Port gives us the port as an int.
var _ = func() int {
	mr, _ := miniredis.Run()
	if mr == nil {
		return 0
	}
	mr.Close()
	return mr.Server().Addr().Port
}

// Ensure goredis is imported (used inside redis.Adapter internals).
var _ = goredis.NewClient

func TestQuestionDeadlinesAndCompletionTimerAgree(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	game := service.GameQuestions{Questions: []service.Question{{Id: "q1"}, {Id: "q2"}, {Id: "q3"}}}
	require.NoError(t, repo.saveQuestionByGameId(ctx, "timed-game", game))
	before := time.Now()
	duration, err := repo.SetValidAnswerTimeForQuestions(ctx, "timed-game")
	require.NoError(t, err)
	assert.Equal(t, 3*repo.Config.ValidAnswerTimeOut, duration)
	stored, err := repo.GetQuestionsByGameId(ctx, "timed-game")
	require.NoError(t, err)
	for i, q := range stored.Questions {
		assert.WithinDuration(t, before.Add(time.Duration(i+1)*repo.Config.ValidAnswerTimeOut), q.ValidAnswerTime, time.Second)
	}
}

func TestConcurrentReadyPlayersAreNotLost(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	for round := 0; round < 10; round++ {
		gameID := fmt.Sprintf("concurrent-%d", round)
		count := 2
		_, err := repo.UpsertReadyPlayer(ctx, gameID, nil, &count)
		require.NoError(t, err)
		start := make(chan struct{})
		results := make(chan error, 2)
		var wg sync.WaitGroup
		for id := 1; id <= 2; id++ {
			wg.Add(1)
			go func(player int) {
				defer wg.Done()
				<-start
				_, err := repo.UpsertReadyPlayer(ctx, gameID, &player, nil)
				results <- err
			}(id)
		}
		close(start)
		wg.Wait()
		close(results)
		for err := range results {
			require.NoError(t, err)
		}
		value, err := repo.redisDB.Get(ctx, GameStatusPrefix+gameID)
		require.NoError(t, err)
		var stored gameStatus
		require.NoError(t, json.Unmarshal([]byte(value), &stored))
		assert.ElementsMatch(t, []int{1, 2}, stored.Players)
	}
}

func TestReadyBeforeRosterInitializationIsRetained(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	player, count := 1, 2
	ready, err := repo.UpsertReadyPlayer(ctx, "early-ready", &player, nil)
	require.NoError(t, err)
	assert.False(t, ready)
	ready, err = repo.UpsertReadyPlayer(ctx, "early-ready", nil, &count)
	require.NoError(t, err)
	assert.False(t, ready)
	player = 2
	ready, err = repo.UpsertReadyPlayer(ctx, "early-ready", &player, nil)
	require.NoError(t, err)
	assert.True(t, ready)
}
