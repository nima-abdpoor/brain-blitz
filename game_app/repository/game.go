package repository

import (
	"BrainBlitz.com/game/game_app/service"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
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
		return "", richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		//todo check the possibility of conversion.
		return result.InsertedID.(primitive.ObjectID).Hex(), nil
	}
}
