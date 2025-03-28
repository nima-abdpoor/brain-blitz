package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/contract/match/golang"
	errApp "BrainBlitz.com/game/pkg/err_app"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"net"
	"strconv"
)

const (
	MatchCreated = "match_created"
)

type IdToConnection map[uint64]net.Conn

type Config struct{}

type Repository interface {
	CreateMatch(ctx context.Context, game Game) (string, error)
}

type Service struct {
	config      Config
	repository  Repository
	webSocket   websocket.WebSocket
	connections IdToConnection
	logger      logger.SlogAdapter
}

func NewService(config Config, repo Repository, ws websocket.WebSocket, logger logger.SlogAdapter) Service {
	return Service{
		config:      config,
		repository:  repo,
		webSocket:   ws,
		logger:      logger,
		connections: IdToConnection{},
	}
}

func (svc Service) ProcessGame(ctx echo.Context, request ProcessGameRequest) (ProcessGameResponse, error) {
	const op = "game.processGame"

	connection, _, _, err := svc.webSocket.Upgrade(ctx.Request(), ctx.Response())
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

		fmt.Println(op, "game created!", result)
	}

	for _, createdMatch := range createdMatches {
		err = svc.writeMessage(createdMatch.UserId, MatchCreated)
		if err != nil {
			svc.logger.Error(op, "error writing message", "userId", fmt.Sprintf("%v", createdMatch.UserId), "message", MatchCreated, slog.String("error", err.Error()))
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
