package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

const (
	QuestionsPrefix      = "game_questions_"
	GameUserStatusPrefix = "game_user_status_"
	GameStatusPrefix     = "game_game_status_"
)

type Config struct {
	QuestionsTimeOut  time.Duration `koanf:"questions_timeout"`
	GameStatusTimeOut time.Duration `koanf:"game_status_timeout"`
}

type GameRepository struct {
	Config  Config
	Logger  logger.SlogAdapter
	MongoDB *mongo.Database
	redisDB *redis.Adapter
}

func NewGameRepository(config Config, logger logger.SlogAdapter, db *mongo.Database, redis *redis.Adapter) service.Repository {
	return GameRepository{
		Config:  config,
		Logger:  logger,
		MongoDB: db,
		redisDB: redis,
	}
}

func (m GameRepository) CreateGame(ctx context.Context, game service.Game) (string, error) {
	const op = "game.CreateGame"
	questions, err := m.GetQuestionsByMatchId(ctx, game.MatchId)
	if err != nil {
		m.Logger.Error(op, "failed to get questions by matchId", err.Error())
	} else {
		game.Question = &questions
	}

	coll := m.MongoDB.DB.Collection("game")
	if result, err := coll.InsertOne(ctx, game); err != nil {
		return "", errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
			"message": "Can not create game",
			"data":    fmt.Sprint(game),
		}, m.Logger)
	} else {
		//todo check the possibility of conversion.
		return result.InsertedID.(primitive.ObjectID).Hex(), nil
	}
}

func (m GameRepository) SaveQuestionsByMatchId(ctx context.Context, matchId string, questions []service.Question) error {
	op := "game.SaveQuestionsByMatchId"
	res, err := json.Marshal(questions)
	if err != nil {
		return err
	}

	filter := bson.M{"matchid": matchId}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"question":   questions,
		},
	}
	coll := m.MongoDB.DB.Collection("game")
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		m.Logger.Error(op, "save questions by matchId in mongoDB", err.Error())
	} else {
		if result.MatchedCount == 0 {
			m.Logger.Error(op, "message", fmt.Sprintf("no game found with matchId %s", matchId))
		}
	}

	return m.redisDB.Set(ctx, QuestionsPrefix+matchId, res, m.Config.QuestionsTimeOut)
}

func (m GameRepository) GetQuestionsByMatchId(ctx context.Context, matchId string) ([]service.Question, error) {
	value, err := m.redisDB.Get(ctx, QuestionsPrefix+matchId)
	if err != nil {
		return nil, err
	}

	var questions []service.Question
	err = json.Unmarshal([]byte(value), &questions)
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (m GameRepository) UpsertUserStatus(ctx context.Context, userId uint64, status service.GameStatus) error {
	return m.redisDB.Set(ctx, GameUserStatusPrefix+strconv.FormatUint(userId, 10), string(status), m.Config.GameStatusTimeOut)
}

func (m GameRepository) GetUserStatus(ctx context.Context, userId uint64) (service.GameStatus, error) {
	value, err := m.redisDB.Get(ctx, GameUserStatusPrefix+strconv.FormatUint(userId, 10))
	if err != nil {
		return service.GameStatusUnknown, err
	}
	return service.MapToGameStatus(value), nil
}

func (m GameRepository) UpsertReadyPlayer(ctx context.Context, gameId string, playerId, numberOfPlayers *int) (bool, error) {
	value, err := m.redisDB.Get(ctx, GameStatusPrefix+gameId)
	if err != nil {
		var players = 2
		if numberOfPlayers != nil {
			players = *numberOfPlayers
		}
		gameStatusJson, err := json.Marshal(gameStatus{
			ExpectedNumberOfPlayers: players,
			Players:                 []int{},
		})
		if err != nil {
			return false, err
		}
		err = m.redisDB.Set(ctx, GameStatusPrefix+gameId, gameStatusJson, m.Config.GameStatusTimeOut)
		if err != nil {
			return false, err
		}

		return false, nil
	}

	var gs gameStatus
	if err = json.Unmarshal([]byte(value), &gs); err != nil {
		return false, err
	}

	if gs.ExpectedNumberOfPlayers-len(gs.Players) == 0 ||
		gs.ExpectedNumberOfPlayers-len(gs.Players) == 1 {
		return true, nil
	}

	for _, id := range gs.Players {
		if id == *playerId {
			return false, fmt.Errorf("player %v is already member of ready players", playerId)
		}
	}

	if playerId != nil {
		gs.Players = append(gs.Players, *playerId)
	}

	gameStatusJson, err := json.Marshal(gs)
	if err != nil {
		return false, err
	}

	err = m.redisDB.Set(ctx, GameStatusPrefix+gameId, gameStatusJson, m.Config.GameStatusTimeOut)
	if err != nil {
		return false, err
	}

	return false, nil
}
