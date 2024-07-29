package matchmanager

import (
	entity "BrainBlitz.com/game/entity/game"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
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
		dock := &MatchCreation{}
		fmt.Println(op, "InsertedID:", result.InsertedID)
		filter := bson.D{{"_id", result.InsertedID}}
		err = coll.FindOne(ctx, filter).Decode(dock)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(op, dock)
	}
	return nil
}
