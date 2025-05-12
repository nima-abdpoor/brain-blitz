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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const (
	QuestionsPrefix = "game_questions_"
)

type Config struct {
	QuestionsTimeOut time.Duration `koanf:"questions_timeout"`
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

func (m GameRepository) CreateMatch(ctx context.Context, game service.Game) (string, error) {
	const op = "game.CreateMatch"

	doc := service.MatchCreation{
		Players:  game.PlayerIDs,
		Category: service.MapFromCategories(game.Category),
		Status:   service.MapToFromGameStatus(game.Status),
	}
	coll := m.MongoDB.DB.Collection("game")
	if result, err := coll.InsertOne(ctx, doc); err != nil {
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
	res, err := json.Marshal(questions)
	if err != nil {
		return err
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
