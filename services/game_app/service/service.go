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

const (
	MatchCreated = "match_created"
)

type IdToConnection map[uint64]net.Conn

type Config struct {
	StoreGameStatusTimeOut          time.Duration `koanf:"store_game_status_time_out"`
	PublishUserToWaitingListTimeOut time.Duration `koanf:"publish_user_to_waiting_list_time_out"`
}

type Repository interface {
	CreateMatch(ctx context.Context, game Game) (string, error)
	SaveQuestionsByMatchId(ctx context.Context, matchId string, questions []Question) error
	GetQuestionsByMatchId(ctx context.Context, matchId string) ([]Question, error)
	UpsertUserStatus(ctx context.Context, userId uint64, status GameStatus) error
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
		result, err := svc.repository.CreateMatch(ctx, Game{
			PlayerIDs: matchedUser.UserId,
			Category:  matchedUser.Category,
		})
		if err != nil {
			svc.logger.Error(op, "error in creating match", slog.String("error", err.Error()))
		} else {
			createdMatches = append(createdMatches, matchedUser)
		}

		svc.logger.Info(op, "game created", result)
	}

	for _, createdMatch := range createdMatches {
		err = svc.writeMessage(createdMatch.UserId, MatchCreated)
		if err != nil {
			svc.logger.Error(op, "error writing message", "userId", fmt.Sprintf("%v", createdMatch.UserId), "message", MatchCreated, slog.String("error", err.Error()))
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
		svc.logger.Error(op, "error in saving questions", slog.String("error", err.Error()))
	}

	return nil
}

func (svc Service) saveUsersGameStatus(userId uint64, status GameStatus) error {
	const op = "game.saveUsersGameStatus"
	ctx, cancel := context.WithTimeout(context.Background(), svc.config.StoreGameStatusTimeOut)
	defer cancel()
	err := svc.repository.UpsertUserStatus(ctx, userId, status)
	if err != nil {
		svc.logger.Error(op, "error in saving user status", slog.String("error", err.Error()))
		return err
	}
	return nil
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
			err = svc.saveUsersGameStatus(id, GameStatusInitialized)
			if err != nil {
				svc.logger.Error(op, "error in storing users GameStatus", slog.String("error", err.Error()))
				response = ProcessGameMessageResponse{
					Success: false,
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
			fmt.Println("ready")
			questions, err := svc.repository.GetQuestionsByMatchId(context.Background(), req.MatchId)
			if err != nil {
				return err
			}

			jsonQuestions, err := json.Marshal(questions)
			if err != nil {
				return err
			}

			err = svc.webSocket.WriteServerData(*conn, code, string(jsonQuestions))
			if err != nil {
				return err
			}
		}
	case CommandGetCategories:
		{
			fmt.Println("get categories")
		}
	case CommandUnknownCommand:
		{
			fmt.Println("unknown")
		}
	}
	return nil
}

func (svc Service) writeMessage(ids []uint64, msg string) error {
	const op = "game.service.writeMessage"

	for _, id := range ids {
		if connection, exists := svc.connections[id]; !exists {
			return fmt.Errorf("id: %d not found", id)
		} else {
			err := svc.webSocket.WriteServerData(connection, websocket.OpText, msg)
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
