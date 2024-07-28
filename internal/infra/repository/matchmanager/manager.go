package matchmanager

import (
	entity "BrainBlitz.com/game/entity/game"
	"context"
	"fmt"
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

func (m MatchManager) CreateMatch(ctx context.Context, game entity.Game) error {
	const op = "matchmanager.CreateMatch"

	doc := MatchCreation{
		UserId:   game.PlayerIDs,
		Category: entity.MapFromCategory(game.Category),
		Status:   entity.MapToFromGameStatus(game.Status),
	}
	coll := m.DB.Collection("match")
	if result, err := coll.InsertOne(ctx, doc); err != nil {
		return err
	} else {
		fmt.Println(op, result)
	}
	return nil
}
