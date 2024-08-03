package matchmanager

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MatchManager struct {
	DB *mongo.Database
}

func New(db *mongo.Database) MatchManager {
	return MatchManager{
		DB: db,
	}
}

func (m MatchManager) CreateMatch(ctx context.Context, game entity.Game) (string, error) {
	const op = "matchmanager.CreateMatch"

	doc := MatchCreation{
		UserId:   game.PlayerIDs,
		Category: entity.MapFromCategory(game.Category),
		Status:   entity.MapToFromGameStatus(game.Status),
	}
	coll := m.DB.Collection("match")
	if result, err := coll.InsertOne(ctx, doc); err != nil {
		return "", richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		//todo check the possibility of conversion.
		return result.InsertedID.(primitive.ObjectID).Hex(), nil
	}
}
