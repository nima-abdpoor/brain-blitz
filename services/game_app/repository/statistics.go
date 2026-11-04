package repository

import (
	"context"

	"BrainBlitz.com/game/pkg/playerstats"
	"BrainBlitz.com/game/services/game_app/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aggregate once per roster member per completed game. Starting from the roster
// includes silent players and zero-score games; unfinished answers never count.
func statisticsPipeline() []bson.M {
	score := bson.M{"$ifNull": bson.A{"$answer.points", 0}}
	top := bson.M{"$eq": bson.A{score, "$topScore"}}
	tied := bson.M{"$gt": bson.A{"$topCount", 1}}
	countIf := func(condition interface{}) bson.M { return bson.M{"$sum": bson.M{"$cond": bson.A{condition, 1, 0}}} }
	return []bson.M{
		{"$match": bson.M{"status": service.GameStatusFinished}},
		{"$project": bson.M{"players": 1}},
		{"$lookup": bson.M{"from": "player_answers", "let": bson.M{"gameID": bson.M{"$toString": "$_id"}}, "pipeline": bson.A{
			bson.M{"$match": bson.M{"$expr": bson.M{"$eq": bson.A{"$game_id", "$$gameID"}}}},
			bson.M{"$group": bson.M{"_id": "$player_id", "points": bson.M{"$sum": "$point"}, "answers": bson.M{"$sum": 1}, "correct": countIf(bson.M{"$gt": bson.A{"$point", 0}})}},
		}, "as": "answers"}},
		{"$set": bson.M{"scores": bson.M{"$map": bson.M{"input": "$players", "as": "player", "in": bson.M{
			"id": "$$player", "answer": bson.M{"$arrayElemAt": bson.A{bson.M{"$filter": bson.M{"input": "$answers", "as": "a", "cond": bson.M{"$eq": bson.A{"$$a._id", bson.M{"$toString": "$$player"}}}}}, 0}},
		}}}}},
		{"$set": bson.M{"topScore": bson.M{"$ifNull": bson.A{bson.M{"$max": "$scores.answer.points"}, 0}}}},
		{"$set": bson.M{"topCount": bson.M{"$size": bson.M{"$filter": bson.M{"input": "$scores", "as": "s", "cond": bson.M{"$eq": bson.A{bson.M{"$ifNull": bson.A{"$$s.answer.points", 0}}, "$topScore"}}}}}}},
		{"$unwind": "$scores"},
		{"$set": bson.M{"answer": "$scores.answer"}},
		{"$group": bson.M{
			"_id": "$scores.id", "gamesPlayed": bson.M{"$sum": 1},
			"wins":        countIf(bson.M{"$and": bson.A{top, bson.M{"$not": bson.A{tied}}}}),
			"draws":       countIf(bson.M{"$and": bson.A{top, tied}}),
			"losses":      countIf(bson.M{"$gt": bson.A{"$topScore", score}}),
			"totalPoints": bson.M{"$sum": score}, "bestScore": bson.M{"$max": score},
			"answersSubmitted": bson.M{"$sum": "$answer.answers"}, "correctAnswers": bson.M{"$sum": "$answer.correct"},
		}},
	}
}

func (m GameRepository) PlayerStatistics(ctx context.Context, userID uint64) (playerstats.Statistics, error) {
	pipeline := statisticsPipeline()
	// A competition rank is one plus the number of players with more points.
	// Do this inside Mongo so the API returns one document, not the global roster.
	pipeline = append(pipeline,
		bson.M{"$setWindowFields": bson.M{"sortBy": bson.M{"totalPoints": -1}, "output": bson.M{"rank": bson.M{"$rank": bson.M{}}, "rankedPlayers": bson.M{"$count": bson.M{}, "window": bson.M{"documents": bson.A{"unbounded", "unbounded"}}}}}},
		bson.M{"$facet": bson.M{"player": bson.A{bson.M{"$match": bson.M{"_id": userID}}}, "population": bson.A{bson.M{"$count": "count"}}}},
	)
	cursor, err := m.MongoDB.DB.Collection("game").Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return playerstats.Statistics{}, err
	}
	defer cursor.Close(ctx)
	var rows []struct {
		Player     []playerstats.Statistics `bson:"player"`
		Population []struct {
			Count int64 `bson:"count"`
		} `bson:"population"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return playerstats.Statistics{}, err
	}
	result := playerstats.Statistics{}
	if len(rows) > 0 {
		if len(rows[0].Population) > 0 {
			result.RankedPlayers = rows[0].Population[0].Count
		}
		if len(rows[0].Player) > 0 {
			p := rows[0].Player[0]
			result = p
		}
	}
	result.CalculateRates()
	return result, nil
}
