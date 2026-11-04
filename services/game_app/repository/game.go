package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/services/game_app/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const (
	QuestionsPrefix      = "game_questions_"
	GameUserStatusPrefix = "game_user_status_"
	GameStatusPrefix     = "game_game_status_"
)

type Config struct {
	QuestionsTimeOut   time.Duration `koanf:"questions_timeout"`
	GameStatusTimeOut  time.Duration `koanf:"game_status_timeout"`
	ValidAnswerTimeOut time.Duration `koanf:"valid_answer_timeout"`
	ScoreConfig        ScoreConfig   `koanf:"score"`
}

type ScoreConfig struct {
	BaseScore     int           `koanf:"base_score"`
	MaxBonus      int           `koanf:"max_bonus"`
	BonusDeadline time.Duration `koanf:"bonus_deadline"`
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
		m.Logger.Debug(op, "questions not delivered yet", "error", err.Error())
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
		gameID := result.InsertedID.(primitive.ObjectID).Hex()
		cached, cacheErr := m.GetQuestionsByMatchId(ctx, game.MatchId)
		if cacheErr == nil && len(cached.Questions) > 0 {
			if err := m.SaveQuestionsByMatchId(ctx, game.MatchId, cached.Questions); err != nil {
				return "", err
			}
		}
		return gameID, nil
	}
}

func (m GameRepository) GetGame(ctx context.Context, gameId string) (service.Game, error) {
	op := "game.GetGame"

	coll := m.MongoDB.DB.Collection("game")
	objectId, err := primitive.ObjectIDFromHex(gameId)
	if err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in converting gameId to objectId %s", gameId), "error", err.Error())
		return service.Game{}, err
	}
	filter := bson.M{"_id": objectId}
	var game service.Game

	if err := coll.FindOne(ctx, filter).Decode(&game); err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in getting game record of %s", filter), "error", err.Error())
		return game, err
	}

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
			m.Logger.Debug(op, "questions arrived before game creation", "matchId", matchId)
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
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return nil
		}
		return err
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
	if errors.Is(err, goredis.Nil) {
		return service.GameStatusUnknown, nil
	}
	if err != nil {
		return service.GameStatusUnknown, err
	}
	return service.MapToGameStatus(value), nil
}

func (m GameRepository) UpsertReadyPlayer(ctx context.Context, gameId string, playerId, numberOfPlayers *int) (bool, error) {
	key := GameStatusPrefix + gameId
	client := m.redisDB.Client()
	for attempt := 0; attempt < 8; attempt++ {
		ready := false
		err := client.Watch(ctx, func(tx *goredis.Tx) error {
			gs := gameStatus{ExpectedNumberOfPlayers: 2, Players: []int{}}
			value, err := tx.Get(ctx, key).Result()
			if err == nil {
				if err := json.Unmarshal([]byte(value), &gs); err != nil {
					return err
				}
			} else if !errors.Is(err, goredis.Nil) {
				return err
			}
			if numberOfPlayers != nil {
				gs.ExpectedNumberOfPlayers = *numberOfPlayers
			}
			if len(gs.Players) == gs.ExpectedNumberOfPlayers {
				ready = true
				return nil
			}
			if playerId != nil {
				for _, id := range gs.Players {
					if id == *playerId {
						return fmt.Errorf("player %v is already member of ready players", playerId)
					}
				}
				gs.Players = append(gs.Players, *playerId)
			}
			data, err := json.Marshal(gs)
			if err != nil {
				return err
			}
			_, err = tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
				pipe.Set(ctx, key, data, m.Config.GameStatusTimeOut)
				return nil
			})
			ready = len(gs.Players) == gs.ExpectedNumberOfPlayers
			return err
		}, key)
		if errors.Is(err, goredis.TxFailedErr) {
			continue
		}
		if err != nil {
			return false, err
		}
		return ready, nil
	}
	return false, fmt.Errorf("ready-player update conflicted repeatedly for game %s", gameId)
}

func (m GameRepository) IncreaseGameQuestionCurrentIndex(ctx context.Context, gameId string) error {
	op := "game.IncreaseGameQuestionCurrentIndex"

	gameQuestions, err := m.GetQuestionsByGameId(ctx, gameId)
	if err != nil {
		m.Logger.Error(op, "get questions by gameId", "gameId", gameId, "error", err.Error())
		return err
	}

	gameQuestions.Questions[gameQuestions.CurrentQuestionIndex].ValidAnswerTime = time.Now().UTC().Add(m.Config.ValidAnswerTimeOut)

	gameQuestions.CurrentQuestionIndex++

	err = m.saveQuestionByGameId(ctx, gameId, gameQuestions)
	if err != nil {
		m.Logger.Error(op, "error in saving game questions", "gameId", gameId, "error", err.Error())
		return err
	}

	return nil
}

func (m GameRepository) SetValidAnswerTimeForQuestions(ctx context.Context, gameId string) (finalTTL time.Duration, err error) {
	op := "game.SetValidAnswerTimeForQuestions"

	gameQuestions, err := m.GetQuestionsByGameId(ctx, gameId)
	if err != nil {
		m.Logger.Error(op, "get questions by gameId", "gameId", gameId, "error", err.Error())
		return 0, err
	}

	startedAt := time.Now().UTC()
	for i := range gameQuestions.Questions {
		ttl := m.Config.ValidAnswerTimeOut * time.Duration(i+1)
		finalTTL = ttl
		gameQuestions.Questions[i].ValidAnswerTime = startedAt.Add(ttl)
	}

	err = m.saveQuestionByGameId(ctx, gameId, gameQuestions)
	if err != nil {
		m.Logger.Error(op, "error in saving game questions", "gameId", gameId, "error", err.Error())
		return 0, err
	}

	return finalTTL, err
}

func (m GameRepository) saveQuestionByGameId(ctx context.Context, gameId string, gameQuestions service.GameQuestions) error {
	gs, err := json.Marshal(&gameQuestions)
	if err != nil {
		return err
	}

	return m.redisDB.Set(ctx, QuestionsPrefix+gameId, gs, m.Config.QuestionsTimeOut)
}

func (m GameRepository) SavePlayerAnswer(ctx context.Context, gameId string, playerAnswer service.PlayerAnswer) error {
	op := "game.SavePlayerAnswer"

	gameQuestion, err := m.GetQuestionsByGameId(context.Background(), gameId)
	if err != nil {
		m.Logger.Error(op, "get questions by gameId", "gameId", gameId, "error", err.Error())
		return err
	}

	if gameQuestion.RoundBased && (gameQuestion.Completed || gameQuestion.CurrentQuestionIndex >= len(gameQuestion.Questions) || gameQuestion.Questions[gameQuestion.CurrentQuestionIndex].Id != playerAnswer.QuestionIDs) {
		return fmt.Errorf("question is no longer active")
	}
	if gameQuestion.RoundBased && !time.Now().Before(gameQuestion.Questions[gameQuestion.CurrentQuestionIndex].ValidAnswerTime) {
		return fmt.Errorf("question deadline has passed")
	}
	questionIdToTTL := make(map[string]time.Time)
	questionIdFound := false
	for _, question := range gameQuestion.Questions {
		if !questionIdFound && playerAnswer.QuestionIDs == question.Id {
			playerAnswer.CorrectChoice = question.CorrectAnswer
			playerAnswer.Options = question.Choices
			playerAnswer.Category = question.Category
			playerAnswer.ValidTimeToAnswer = question.ValidAnswerTime
			timeDiff := time.Duration(question.ValidAnswerTime.Sub(playerAnswer.AnswerTime).Seconds())
			if question.ValidAnswerTime.Sub(playerAnswer.AnswerTime).Milliseconds() > m.Config.ValidAnswerTimeOut.Milliseconds() {
				m.Logger.Info(op,
					"message", "subtract of player answerTime and the validAnswerTime is greater than allowable Question ttl",
					"validTimeAnswer", playerAnswer.ValidTimeToAnswer,
					"playerTimeAnswer", playerAnswer.AnswerTime,
					"timeDiff", timeDiff.Milliseconds(),
					"question ttl", m.Config.ValidAnswerTimeOut.Milliseconds(),
				)

				return errApp.New(op,
					"INVALID_ANSWER",
					"answered to question quickly",
					http.StatusBadRequest,
					codes.InvalidArgument,
					nil,
					nil,
				)
			}
			isCorrect := playerAnswer.PlayerChoice == question.CorrectAnswer
			score := m.CalculateScore(isCorrect, playerAnswer.AnswerTime, question.ValidAnswerTime)
			playerAnswer.TimeDiff = timeDiff
			playerAnswer.Point = score
			questionIdFound = true

			m.Logger.Info(op+"_score calculated",
				"gameId", gameId,
				"playerId", playerAnswer.PlayerID,
				"validTimeAnswer", playerAnswer.ValidTimeToAnswer,
				"playerTimeAnswer", playerAnswer.AnswerTime,
				"timeDiff", timeDiff,
				"isCorrect", isCorrect,
				"point", score,
			)
		}
		questionIdToTTL[question.Id] = question.ValidAnswerTime
	}

	coll := m.MongoDB.DB.Collection("player_answers")

	storedPlayerAnswerFilter := bson.M{"game_id": gameId, "player_id": playerAnswer.PlayerID, "question_id": playerAnswer.QuestionIDs}
	var storedPlayerAnswer service.PlayerAnswer
	err = coll.FindOne(context.Background(), storedPlayerAnswerFilter).Decode(&storedPlayerAnswer)
	if err != nil && !errors.Is(err, mongodriver.ErrNoDocuments) {
		return fmt.Errorf("checking duplicate answer: %w", err)
	}
	if storedPlayerAnswer.PlayerID == playerAnswer.PlayerID &&
		storedPlayerAnswer.GameId == playerAnswer.GameId &&
		storedPlayerAnswer.QuestionIDs == playerAnswer.QuestionIDs {
		return fmt.Errorf("player %v already answered this question %s", playerAnswer.PlayerID, playerAnswer.QuestionIDs)
	}

	_, err = coll.InsertOne(context.Background(), playerAnswer)
	if err != nil {
		return errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
			"message": "Can not insert player answer",
			"data":    fmt.Sprint(playerAnswer),
		}, m.Logger)
	}

	return nil
}

func (m GameRepository) GetLeaderBoard(ctx context.Context, gameId string) (service.LeaderBoard, error) {
	op := "game.GetLeaderBoard"

	var leaderBoard service.LeaderBoard
	var playerAnswerResult []service.PlayerAnswer

	gameCollection := m.MongoDB.DB.Collection("game")
	objectId, err := primitive.ObjectIDFromHex(gameId)
	if err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in converting gameId to objectId %s", gameId), "error", err.Error())
		return leaderBoard, err
	}
	gameFilter := bson.M{"_id": objectId}
	var storedGame service.Game

	if err = gameCollection.FindOne(context.Background(), gameFilter).Decode(&storedGame); err != nil {
		m.Logger.Error(op, fmt.Sprintf("error in getting game record of %s", gameFilter), "error", err.Error())
		return leaderBoard, err
	}

	playerAnswersCollection := m.MongoDB.DB.Collection("player_answers")
	filter := bson.M{
		"game_id": gameId,
	}

	cursor, err := playerAnswersCollection.Find(context.Background(), filter)
	if err != nil {
		m.Logger.Error(op, "find all player answers", "filter", fmt.Sprint(filter), "error", err.Error())
		return leaderBoard, err
	}
	//defer cursor.Close()

	if err := cursor.All(ctx, &playerAnswerResult); err != nil {
		m.Logger.Error(op, "decode player answers", "error", err.Error())
		return leaderBoard, err
	}

	var playerPoint []service.PlayerPoint
	playerToPoint := make(map[string]*service.PlayerPoint)

	for _, ans := range playerAnswerResult {
		isCorrect := ans.TimeDiff > 0 && ans.Point > 0
		if player, exists := playerToPoint[ans.PlayerID]; exists {
			player.Point += ans.Point

			player.QuestionCorrectness = append(player.QuestionCorrectness, service.QuestionCorrectness{
				QuestionId:    ans.QuestionIDs,
				PlayerChoice:  ans.PlayerChoice,
				CorrectChoice: ans.CorrectChoice,
				IsCorrect:     isCorrect,
			})
		} else {
			playerToPoint[ans.PlayerID] = &service.PlayerPoint{
				PlayerId: ans.PlayerID,
				Point:    ans.Point,
				QuestionCorrectness: []service.QuestionCorrectness{{
					QuestionId:    ans.QuestionIDs,
					PlayerChoice:  ans.PlayerChoice,
					CorrectChoice: ans.CorrectChoice,
					IsCorrect:     isCorrect,
				},
				},
			}
		}
	}

	for _, player := range storedGame.Players {
		playerId := strconv.FormatInt(int64(player), 10)
		if _, exists := playerToPoint[playerId]; !exists {
			playerToPoint[playerId] = &service.PlayerPoint{
				PlayerId:            playerId,
				Point:               0,
				QuestionCorrectness: []service.QuestionCorrectness{},
			}
		}
	}

	for player, point := range playerToPoint {
		playerPoint = append(playerPoint, service.PlayerPoint{
			PlayerId:            player,
			Point:               point.Point,
			QuestionCorrectness: point.QuestionCorrectness,
		})
	}

	sort.Slice(playerPoint, func(i, j int) bool {
		return playerPoint[i].Point > playerPoint[j].Point
	})

	leaderBoard = service.LeaderBoard{
		GameId:       gameId,
		PlayersPoint: playerPoint,
	}

	return leaderBoard, nil
}

func (m GameRepository) CalculateScore(isCorrect bool, answerTime, validAnswerTime time.Time) int {
	if !isCorrect || answerTime.After(validAnswerTime) {
		return 0
	}

	timeDiff := validAnswerTime.Sub(answerTime).Seconds()

	var bonus int
	if timeDiff >= m.Config.ScoreConfig.BonusDeadline.Seconds() {
		bonus = m.Config.ScoreConfig.MaxBonus
	} else {
		bonus = int(float64(m.Config.ScoreConfig.MaxBonus) * (timeDiff / m.Config.ScoreConfig.BonusDeadline.Seconds()))
	}

	return m.Config.ScoreConfig.BaseScore + bonus
}

func (m GameRepository) FinishGame(ctx context.Context, gameId string) error {
	objectID, err := primitive.ObjectIDFromHex(gameId)
	if err != nil {
		return err
	}
	_, err = m.MongoDB.DB.Collection("game").UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{
		"$set": bson.M{"status": service.GameStatusFinished, "updated_at": time.Now().UTC()},
	})
	return err
}
