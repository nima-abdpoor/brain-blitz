package repository

import (
	"context"
	"os"
	"testing"
	"time"

	dbpkg "BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/services/game_app/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Uses a disposable database; never writes the application's game history.
func TestPlayerStatisticsMongo(t *testing.T) {
	uri := os.Getenv("STATISTICS_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("set STATISTICS_TEST_MONGO_URI to run Mongo statistics regression")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database("statistics_test_" + primitive.NewObjectID().Hex())
	defer db.Drop(context.Background())
	repo := GameRepository{MongoDB: &dbpkg.Database{DB: db}}
	insert := func(status service.GameStatus, players []uint64, points map[string][]int) {
		t.Helper()
		id := primitive.NewObjectID()
		if _, err := db.Collection("game").InsertOne(ctx, bson.M{"_id": id, "status": status, "players": players}); err != nil {
			t.Fatal(err)
		}
		for player, scores := range points {
			for i, score := range scores {
				_, err := db.Collection("player_answers").InsertOne(ctx, bson.M{"game_id": id.Hex(), "player_id": player, "question_id": i, "point": score})
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	insert(service.GameStatusFinished, []uint64{1, 2}, map[string][]int{"1": {10, 5}, "2": {5, 0}})
	insert(service.GameStatusFinished, []uint64{1, 2}, map[string][]int{"1": {0}, "2": {10}})
	insert(service.GameStatusFinished, []uint64{1, 2}, nil)                                    // Both silent: a zero-score draw.
	insert(service.GameStatusFinished, []uint64{3, 4}, map[string][]int{"3": {15}, "4": {15}}) // Tied global scores.
	insert(service.GameStatusStarted, []uint64{1, 2}, map[string][]int{"1": {999}})
	one, err := repo.PlayerStatistics(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if one.GamesPlayed != 3 || one.Wins != 1 || one.Losses != 1 || one.Draws != 1 || one.TotalPoints != 15 || one.BestScore != 15 || one.AnswersSubmitted != 3 || one.CorrectAnswers != 2 || one.Rank != 1 || one.RankedPlayers != 4 || one.AveragePoints != 5 {
		t.Fatalf("unexpected stats: %+v", one)
	}
	two, err := repo.PlayerStatistics(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if two.Rank != 1 || two.TotalPoints != 15 {
		t.Fatalf("ties must share rank: %+v", two)
	}
	insert(service.GameStatusFinished, []uint64{3, 4}, map[string][]int{"3": {1}})
	one, err = repo.PlayerStatistics(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if one.Rank != 2 {
		t.Fatalf("rank must update after a completed game: %+v", one)
	}
	fresh, err := repo.PlayerStatistics(ctx, 99)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Rank != 0 || fresh.GamesPlayed != 0 || fresh.RankedPlayers != 4 {
		t.Fatalf("new player: %+v", fresh)
	}
}
