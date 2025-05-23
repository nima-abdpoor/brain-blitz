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
		game.Question = &questions.Questions
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

func (m GameRepository) GetGame(ctx context.Context, gameId string) (service.Game, error) {
	op := "game.GetGame"

	coll := m.MongoDB.DB.Collection("game")
	filter := bson.M{"_id": gameId}
	var game service.Game

	if err := coll.FindOne(ctx, filter).Decode(&game); err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in getting game record of %s", filter), "error", err.Error())
		return game, err
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), m.Config.GameStatusTimeOut)
		defer cancel()

		gs := service.GameQuestions{
			Questions:            *game.Question,
			Players:              game.Players,
			CurrentQuestionIndex: 0,
		}
		err := m.saveQuestionByGameId(cacheCtx, gameId, gs)
		if err != nil {
			m.Logger.Error(op, fmt.Sprintf("error in caching game questions %s", gameId), "error", err.Error())
		}
	}()

	return game, nil
}

func (m GameRepository) SaveQuestionsByMatchId(ctx context.Context, matchId string, questions []service.Question) error {
	op := "game.SaveQuestionsByMatchId"

	gQ := service.GameQuestions{
		Questions:            questions,
		CurrentQuestionIndex: 0,
	}
	res, err := json.Marshal(gQ)
	if err != nil {
		return err
	}

	filter := bson.M{"match_id": matchId}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
			"questions":  questions,
		},
	}
	coll := m.MongoDB.DB.Collection("game")
	updateResult, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		m.Logger.Error(op, "save questions by matchId in mongoDB", err.Error())

		redisErr := m.redisDB.Set(ctx, QuestionsPrefix+matchId, res, m.Config.QuestionsTimeOut)
		if redisErr != nil {
			m.Logger.Error(op, "save questions by matchId in redis", err.Error())
		}
	} else {
		if updateResult.MatchedCount == 0 {
			m.Logger.Error(op, "message", fmt.Sprintf("no game found with matchId %s", matchId))
			redisErr := m.redisDB.Set(ctx, QuestionsPrefix+matchId, res, m.Config.QuestionsTimeOut)
			if redisErr != nil {
				m.Logger.Error(op, "save questions by matchId in redis", err.Error())
			}
		}
	}

	var selectResult struct {
		ID      primitive.ObjectID `bson:"_id"`
		Players []uint64           `bson:"players"`
	}
	if err := coll.FindOne(ctx, filter).Decode(&selectResult); err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in getting game record of %s", filter), "error", err.Error())
	}

	gQ.Players = selectResult.Players
	gqJson, err := json.Marshal(gQ)
	if err != nil {
		return err
	}

	return m.redisDB.Set(ctx, QuestionsPrefix+selectResult.ID.Hex(), gqJson, m.Config.QuestionsTimeOut)
}

func (m GameRepository) GetQuestionsByGameId(ctx context.Context, gameId string) (service.GameQuestions, error) {
	var questions service.GameQuestions
	value, err := m.redisDB.Get(ctx, QuestionsPrefix+gameId)
	if err != nil {
		return questions, err
	}

	err = json.Unmarshal([]byte(value), &questions)
	if err != nil {
		return questions, err
	}

	return questions, nil
}

func (m GameRepository) GetQuestionsByMatchId(ctx context.Context, matchId string) (service.GameQuestions, error) {
	var questions service.GameQuestions
	value, err := m.redisDB.Get(ctx, QuestionsPrefix+matchId)
	if err != nil {
		return questions, err
	}

	err = json.Unmarshal([]byte(value), &questions)
	if err != nil {
		return questions, err
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

func (m GameRepository) IncreaseGameQuestionCurrentIndex(ctx context.Context, gameId string) error {
	op := "game.IncreaseGameQuestionCurrentIndex"

	gameQuestions, err := m.GetQuestionsByGameId(ctx, gameId)
	if err != nil {
		m.Logger.Error(op, "get questions by gameId", "gameId", gameId, "error", err.Error())
		return err
	}

	gameQuestions.CurrentQuestionIndex++

	err = m.saveQuestionByGameId(ctx, gameId, gameQuestions)
	if err != nil {
		m.Logger.Error(op, "error in saving game questions", "gameId", gameId, "error", err.Error())
		return err
	}

	return nil
}

func (m GameRepository) saveQuestionByGameId(ctx context.Context, gameId string, gameQuestions service.GameQuestions) error {
	gs, err := json.Marshal(&gameQuestions)
	if err != nil {
		return err
	}

	return m.redisDB.Set(ctx, QuestionsPrefix+gameId, gs, m.Config.QuestionsTimeOut)
}
