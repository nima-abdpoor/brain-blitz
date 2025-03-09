package repository

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/game_app/service"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

type Config struct{}

type GameRepository struct {
	Config  Config
	Logger  *slog.Logger
	MongoDB *mongo.Database
}

func NewGameRepository(config Config, logger *slog.Logger, db *mongo.Database) service.Repository {
	return GameRepository{
		Config:  config,
		Logger:  logger,
		MongoDB: db,
	}
}

func (m GameRepository) CreateMatch(ctx context.Context, game entity.Game) (string, error) {
	const op = "game.CreateMatch"

	doc := service.MatchCreation{
		Players:  game.PlayerIDs,
		Category: entity.MapFromCategories(game.Category),
		Status:   entity.MapToFromGameStatus(game.Status),
	}
	coll := m.MongoDB.DB.Collection("game")
	if result, err := coll.InsertOne(ctx, doc); err != nil {
		return "", richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		//todo check the possibility of conversion.
		return result.InsertedID.(primitive.ObjectID).Hex(), nil
	}
}
