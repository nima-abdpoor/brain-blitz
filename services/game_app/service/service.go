package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/contract/match/golang"
	questionProto "BrainBlitz.com/game/contract/question/golang"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"net"
	"strconv"
	"time"
)

type IdToConnection map[uint64]net.Conn

type Config struct {
	StoreGameStatusTimeOut          time.Duration `koanf:"store_game_status_time_out"`
	PublishUserToWaitingListTimeOut time.Duration `koanf:"publish_user_to_waiting_list_time_out"`
	QuestionInterval                time.Duration `koanf:"question_interval"`
}

const saveUserGameStatusTimeOut = 2 * time.Second
const saveGameStatusTimeOut = 2 * time.Second

type Repository interface {
	CreateGame(ctx context.Context, game Game) (string, error)
	GetGame(ctx context.Context, gameId string) (Game, error)

	SaveQuestionsByMatchId(ctx context.Context, matchId string, questions []Question) error
	GetQuestionsByGameId(ctx context.Context, gameId string) (GameQuestions, error)
	GetQuestionsByMatchId(ctx context.Context, matchId string) (GameQuestions, error)
	IncreaseGameQuestionCurrentIndex(ctx context.Context, gameId string) error

	SavePlayerAnswer(ctx context.Context, gameId string, playerAnswer PlayerAnswer) (LeaderBoard, error)

	UpsertUserStatus(ctx context.Context, userId uint64, status GameStatus) error
	UpsertReadyPlayer(ctx context.Context, gameId string, playerId, numberOfPlayers *int) (bool, error)
	GetUserStatus(ctx context.Context, userId uint64) (GameStatus, error)
}

type Service struct {
	config      Config
	repository  Repository
	webSocket   websocket.WebSocket
	connections IdToConnection
	broker      broker.Broker
	logger      logger.SlogAdapter
}

func NewService(config Config, repo Repository, ws websocket.WebSocket, broker broker.Broker, logger logger.SlogAdapter) Service {
	return Service{
		config:      config,
		repository:  repo,
		webSocket:   ws,
		logger:      logger,
		broker:      broker,
		connections: IdToConnection{},
	}
}

func (svc Service) ProcessGame(ctx echo.Context, request ProcessGameRequest) (ProcessGameResponse, error) {
	const op = "game.processGame"

	connection, rw, _, err := svc.webSocket.Upgrade(ctx.Request(), ctx.Response())
	if err != nil {
		return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
			"message": "error in initializing websocket",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}

	id, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
			"message": "error in converting id to Uint",
			"data":    fmt.Sprint(request),
		}, svc.logger)
	}

	svc.connections[id] = *connection
	switch svc.getUsersGameStatus(id) {
	case GameStatusInitialized:
		{

		}
	case GameStatusPending:
		{

		}
	case GameStatusCreated:
		{

		}
	case GameStatusStarted:
		{

		}
	case GameStatusFinished:
		{
			//err = svc.saveUsersGameStatus(id, GameStatusInitialized)
			//if err != nil {
			//	svc.logger.Error(op, "error in storing users GameStatus", slog.String("error", err.Error()))
			//}
		}
	case GameStatusUnknown:
		{
			categories := svc.getCategories()
			categoriesByte, err := json.Marshal(categories)
			if err != nil {
				return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
					"message": "error in marshaling json of Categories",
					"data":    fmt.Sprint(request),
				}, svc.logger)
			}
			err = svc.webSocket.WriteServerData(svc.connections[id], websocket.OpText, string(categoriesByte))
			if err != nil {
				return ProcessGameResponse{}, errApp.Wrap(op, nil, errApp.ErrInternal, map[string]string{
					"message": "error in returning Categories",
					"data":    fmt.Sprint(request),
				}, svc.logger)
			}
		}
	}

	go func(ctx context.Context, conn *net.Conn, rw *bufio.ReadWriter, userID uint64) {
		defer func() {
			(*conn).Close()
			delete(svc.connections, userID)
			svc.logger.Info("connection closed", "userID", userID)
		}()

		for {
			msg, code, err := svc.webSocket.ReadClientData(rw)
			if err != nil {
				svc.logger.Error("read failed", "userID", userID, "error", err)
				break
			}
			err = svc.readMessage(ctx, id, conn, code, msg)
			if err != nil {
				svc.logger.Error("read failed", "userID", userID, "error", err)
			}
		}
	}(ctx.Request().Context(), connection, rw, id)

	return ProcessGameResponse{}, nil
}

func (svc Service) ConsumeMatchCreated(message []byte, ctx context.Context) error {
	const op = "game.consumeMatchCreated"

	users := &golang.AllMatchedUsers{}
	err := proto.Unmarshal(message, users)
	if err != nil {
		svc.logger.Error(op, "error in unmarshalling match message", slog.String("error", err.Error()))
		return err
	}
	matchedUsers := MapFromProtoMessageToEntity(users)
	createdMatches := make([]MatchedUsers, 0)
	for _, matchedUser := range matchedUsers {
		result, err := svc.repository.CreateGame(ctx, Game{
			Id:        nil,
			Players:   matchedUser.UserId,
			MatchId:   matchedUser.MatchId,
			Category:  matchedUser.Category,
			Status:    GameStatusPending,
			Question:  nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			svc.logger.Error(op, "error in creating match", slog.String("error", err.Error()))
		} else {
			matchedUser.GameId = result
			createdMatches = append(createdMatches, matchedUser)
		}

		svc.logger.Info(op, "game created", result)
	}

	for _, createdMatch := range createdMatches {
		go svc.saveUsersGameStatus(createdMatch.UserId, GameStatusPending)

		numberOfPlayers := len(createdMatch.UserId)
		go svc.saveGameStatus(createdMatch.GameId, nil, &numberOfPlayers)

		msg := ProcessGameMessageResponse{
			Success: true,
			Event:   EventMatchCreated,
			Message: "match created",
			MetaData: ProcessGameMetaDataResponse{
				GameId: createdMatch.GameId,
			},
		}
		err = svc.writeMessage(createdMatch.UserId, msg)
		if err != nil {
			svc.logger.Error(op, "error writing message", "userId", fmt.Sprintf("%v", createdMatch.UserId), "message", EventMatchCreated, slog.String("error", err.Error()))
		}
	}

	return nil
}

func (svc Service) ConsumeQuestions(message []byte, ctx context.Context) error {
	const op = "game.consumeQuestions"

	protoQuestions := &questionProto.Questions{}
	err := proto.Unmarshal(message, protoQuestions)
	if err != nil {
		svc.logger.Error(op, "error in unmarshalling question message", slog.String("error", err.Error()))
		return err
	}
	questions, matchId := MapFromProtoMessageToQuestionsEntity(protoQuestions)

	svc.logger.Info(op, "question list received", questions, "matchId", matchId)

	err = svc.repository.SaveQuestionsByMatchId(ctx, matchId, questions)
	if err != nil {
		svc.logger.Error(op, "error in saving Questions", slog.String("error", err.Error()))
	}

	return nil
}

func (svc Service) saveUsersGameStatus(userId []uint64, status GameStatus) {
	const op = "game.saveUsersGameStatus"

	for _, id := range userId {
		go func(id uint64) {
			upsertUserStatusCtx, cancel := context.WithTimeout(context.Background(), saveUserGameStatusTimeOut)
			defer cancel()

			if err := svc.repository.UpsertUserStatus(upsertUserStatusCtx, id, status); err != nil {
				svc.logger.Error(op, "error in saving user status", "error", err.Error())
			}
		}(id)
	}
}

func (svc Service) saveGameStatus(gameId string, userId *uint64, numberOfPlayers *int) bool {
	const op = "game.saveGameStatus"

	ctx, cancel := context.WithTimeout(context.Background(), saveGameStatusTimeOut)
	defer cancel()

	var id *int
	if userId != nil {
		uId := int(*userId)
		id = &uId
	}
	isGameReady, err := svc.repository.UpsertReadyPlayer(ctx, gameId, id, numberOfPlayers)
	if err != nil {
		svc.logger.Error(op, "error in saving ready player")
	}

	return isGameReady
}

func (svc Service) getUsersGameStatus(userId uint64) GameStatus {
	const op = "game.getUsersGameStatus"
	ctx, cancel := context.WithTimeout(context.Background(), svc.config.StoreGameStatusTimeOut)
	defer cancel()
	status, err := svc.repository.GetUserStatus(ctx, userId)
	if err != nil {
		svc.logger.Error(op, "error in getting user status", slog.String("error", err.Error()))
		return GameStatusUnknown
	}
	return status
}

func (svc Service) readMessage(ctx context.Context, id uint64, conn *net.Conn, code websocket.OpCode, message string) error {
	op := "game.readMessage"
	svc.logger.Info("received message", "code", code, "message", message)

	var req ProcessGameMessageRequest
	var response ProcessGameMessageResponse
	err := json.Unmarshal([]byte(message), &req)
	if err != nil {
		return err
	}

	switch req.Command {
	case CommandAddToWaitingList:
		{
			svc.logger.Info(op, "adding to waiting list")
			if MapToCategory(req.Category) == CategoryTypeUnknown {
				response = ProcessGameMessageResponse{
					Success: false,
					Event:   Error,
					Message: "invalid category",
				}
				addToWaitingListResponse, err := json.Marshal(response)
				if err != nil {
					return err
				}
				err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
				if err != nil {
					return err
				}
			}
			go svc.saveUsersGameStatus([]uint64{id}, GameStatusInitialized)

			brokerCtx, cancel := context.WithTimeout(context.Background(), svc.config.PublishUserToWaitingListTimeOut)
			defer cancel()
			buff, err := proto.Marshal(MapWaitingListRequestToProtoMessage(id, req.Category))
			if err != nil {
				//todo update metrics
				svc.logger.Error(op, "message", "error in marshaling waiting list request message", err.Error())
			}

			err = svc.broker.Publish(brokerCtx, event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED, buff)
			if err != nil {
				svc.logger.Error(op, "error in publishing join request message into broker", slog.String("error", err.Error()))
				response = ProcessGameMessageResponse{
					Success: false,
					Event:   Error,
					Message: "internal server error",
				}
				addToWaitingListResponse, err := json.Marshal(response)
				if err != nil {
					return err
				}
				err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
				if err != nil {
					return err
				}
			}
			response = ProcessGameMessageResponse{
				Success: true,
				Event:   AddedToWaitingList,
				Message: "added to waiting list successfully",
			}
			addToWaitingListResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
			if err != nil {
				return err
			}
		}
	case CommandReady:
		{
			// check user if their status is just initialized
			isGameReady := svc.saveGameStatus(req.GameId, &id, nil)
			if isGameReady {
				err = svc.sendQuestionToPlayer(ctx, req.GameId)
				if err == nil {
					fmt.Println("game completed")
				}
			}
		}
	case CommandGetCategories:
		{
			fmt.Println("get categories")
		}
	case CommandAnswer:
		{
			playerResponse, err := svc.savePlayerAnswer(ctx, id, req.GameAnswer)
			if err != nil {
				svc.logger.Error(op, "error in saving answer of a player", slog.String("error", err.Error()))
			}

			response.Event = AnswerAccepted
			response.Message = "answer accepted"
			response.Success = true
			response.MetaData = ProcessGameMetaDataResponse{
				GameId: req.GameId,
				Answer: playerResponse,
			}
			answerAcceptedJson, err := json.Marshal(response)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(answerAcceptedJson))
			if err != nil {
				return err
			}
		}
	default:
		{
			svc.logger.Error(op, "invalid command", "command", req.Command, req.MatchId)
			response = ProcessGameMessageResponse{
				Success: false,
				Event:   Error,
				Message: "invalid command",
			}
			addToWaitingListResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}
			err = svc.webSocket.WriteServerData(*conn, code, string(addToWaitingListResponse))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (svc Service) sendQuestionToPlayer(ctx context.Context, gameId string) error {
	op := "game.sendQuestionToPlayer"
	svc.logger.Info(op, "start sending questions to players", gameId)

	gameQuestions, err := svc.repository.GetQuestionsByGameId(context.Background(), gameId)
	if err != nil {
		return err
	}

	var questionsResponse []ProcessGameQuestion

	for _, questions := range gameQuestions.Questions {
		questionsResponse = append(questionsResponse, ProcessGameQuestion{
			Id:         questions.Id,
			Content:    questions.Content,
			Choices:    questions.Choices,
			Difficulty: questions.Difficulty,
			TTL:        questions.ValidAnswerTime,
		})
	}

	gameMessageResponse := ProcessGameMessageResponse{
		Success: true,
		Event:   NewQuestion,
		Message: "new question",
		MetaData: ProcessGameMetaDataResponse{
			GameId:    gameId,
			Questions: questionsResponse,
		},
	}

	jsonQuestions, err := json.Marshal(gameMessageResponse)
	if err != nil {
		return err
	}
	for _, playerId := range gameQuestions.Players {
		conn := svc.connections[playerId]
		err = svc.webSocket.WriteServerData(conn, websocket.OpText, string(jsonQuestions))
		if err != nil {
			svc.logger.Error(
				op, "error in sending questions to player",
				"playerId", playerId,
				"gameId", gameId,
				slog.String("error", err.Error()),
			)
		}
	}
	return nil
}

func (svc Service) savePlayerAnswer(ctx context.Context, playerId uint64, answer ProcessGameAnswer) (ProcessGameLeaderBoard, error) {
	op := "game.savePlayerAnswer"

	var leaderBoardResult ProcessGameLeaderBoard

	ps := PlayerAnswer{
		GameId:       answer.GameId,
		QuestionIDs:  answer.QuestionId,
		PlayerID:     strconv.FormatUint(playerId, 10),
		PlayerChoice: answer.Answer,
		AnswerTime:   time.Now(),
	}

	leaderBoard, err := svc.repository.SavePlayerAnswer(ctx, answer.GameId, ps)
	if err != nil {
		svc.logger.Error(op, "error in saving answer of a player", slog.String("error", err.Error()))
		return leaderBoardResult, err
	}

	var playerPoints []ProcessGamePlayerPoint

	for _, playerPoint := range leaderBoard.PlayersPoint {
		var ansResult []ProcessGameAnswerResult
		for _, questionAnswers := range playerPoint.QuestionCorrectness {
			ansResult = append(ansResult, ProcessGameAnswerResult{
				QuestionId:    questionAnswers.QuestionId,
				CorrectAnswer: questionAnswers.CorrectChoice,
				PlayerAnswer:  questionAnswers.PlayerChoice,
				IsCorrect:     questionAnswers.IsCorrect,
			})
		}
		playerPoints = append(playerPoints, ProcessGamePlayerPoint{
			PlayerId: playerPoint.PlayerId,
			Point:    playerPoint.Point,
			Answers:  ansResult,
		})
	}
	leaderBoardResult = ProcessGameLeaderBoard{
		GameId:      answer.GameId,
		PlayerPoint: playerPoints,
	}

	return leaderBoardResult, nil
}

func (svc Service) writeMessage(ids []uint64, msg ProcessGameMessageResponse) error {
	const op = "game.service.writeMessage"

	for _, id := range ids {
		if connection, exists := svc.connections[id]; !exists {
			return fmt.Errorf("id: %d not found", id)
		} else {
			jsonMsg, err := json.Marshal(msg)
			if err != nil {
				svc.logger.Error(op, "message", "error writing message", slog.String("error", err.Error()))
			}
			err = svc.webSocket.WriteServerData(connection, websocket.OpText, string(jsonMsg))
			if err != nil {
				svc.logger.Error(op, "message", "error writing message", slog.String("error", err.Error()))
			}
		}
	}
	return nil
}

func (svc Service) getCategories() GameInitResponse {
	var categories []string

	for _, category := range GetCategories() {
		categories = append(categories, string(category))
	}

	users := []int{2}

	result := GameInitResponse{
		Categories:      categories,
		NumberOfPlayers: users,
	}

	return result
}
