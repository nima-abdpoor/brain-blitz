package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/logger"
	"context"
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
	CreateGame(ctx context.Context, game Game) (string, error)
	GetGame(ctx context.Context, gameId string) (Game, error)

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
	}
}
