package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/playerstats"
	"context"
	"hash/fnv"
	"net"
	"sync"
	"time"
)

type IdToConnection map[uint64]net.Conn

type Config struct {
	StoreGameStatusTimeOut          time.Duration `koanf:"store_game_status_time_out"`
	PublishUserToWaitingListTimeOut time.Duration `koanf:"publish_user_to_waiting_list_time_out"`
	QuestionInterval                time.Duration `koanf:"question_interval"`
}

const saveUserGameStatusTimeOut = 2 * time.Second
const saveGameStatusTimeOut = 2 * time.Second

type Repository interface {
	PlayerStatistics(ctx context.Context, userID uint64) (playerstats.Statistics, error)
	OpenRound(ctx context.Context, gameID string, index int) (GameQuestions, error)
	CreateGame(ctx context.Context, game Game) (string, error)
	GetGame(ctx context.Context, gameId string) (Game, error)
	FinishGame(ctx context.Context, gameId string) error

	SaveQuestionsByMatchId(ctx context.Context, matchId string, questions []Question) error
	GetQuestionsByGameId(ctx context.Context, gameId string) (GameQuestions, error)
	GetQuestionsByMatchId(ctx context.Context, matchId string) (GameQuestions, error)
	IncreaseGameQuestionCurrentIndex(ctx context.Context, gameId string) error
	SetValidAnswerTimeForQuestions(ctx context.Context, gameId string) (time.Duration, error)

	SavePlayerAnswer(ctx context.Context, gameId string, playerAnswer PlayerAnswer) error
	GetLeaderBoard(ctx context.Context, gameId string) (LeaderBoard, error)

	UpsertUserStatus(ctx context.Context, userId uint64, status GameStatus) error
	UpsertReadyPlayer(ctx context.Context, gameId string, playerId, numberOfPlayers *int) (bool, error)
	GetUserStatus(ctx context.Context, userId uint64) (GameStatus, error)
}

type Service struct {
	config        Config
	repository    Repository
	webSocket     websocket.WebSocket
	connections   IdToConnection
	mu            *sync.RWMutex
	roundMu       *[64]sync.Mutex
	playerNames   map[uint64]string
	taskPublisher taskqueue.TaskPublisher
	broker        broker.Broker
	logger        logger.SlogAdapter
}

func NewService(
	config Config,
	repo Repository,
	ws websocket.WebSocket,
	broker broker.Broker,
	taskPublisher taskqueue.TaskPublisher,
	logger logger.SlogAdapter,
) Service {
	return Service{
		config:        config,
		repository:    repo,
		webSocket:     ws,
		logger:        logger,
		broker:        broker,
		taskPublisher: taskPublisher,
		connections:   IdToConnection{},
		mu:            &sync.RWMutex{},
		roundMu:       &[64]sync.Mutex{},
		playerNames:   make(map[uint64]string),
	}
}

// A reader for an old socket must not delete a replacement connection.
func (svc Service) closeConnection(userID uint64, conn net.Conn) {
	if conn == nil {
		return
	}
	_ = conn.Close()
	svc.mu.Lock()
	ours := svc.connections[userID] == conn
	if ours {
		delete(svc.connections, userID)
		svc.leaveQueue(userID)
	}
	svc.mu.Unlock()
}

// Fixed lock stripes serialize each game without retaining a lock per finished
// game, while unrelated matches can accept answers independently.
func (svc Service) roundMutex(gameID string) *sync.Mutex {
	h := fnv.New32a()
	_, _ = h.Write([]byte(gameID))
	return &svc.roundMu[h.Sum32()%uint32(len(svc.roundMu))]
}
