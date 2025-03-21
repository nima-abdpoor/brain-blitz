package repository

import (
	"BrainBlitz.com/game/game_app/service"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Config struct{}

type GameRepository struct {
	Config  Config
	Logger  logger.SlogAdapter
	MongoDB *mongo.Database
}

func NewGameRepository(config Config, logger logger.SlogAdapter, db *mongo.Database) service.Repository {
	return GameRepository{
		Config:  config,
		Logger:  logger,
		MongoDB: db,
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
