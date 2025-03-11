package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net"
	"strconv"
)

type IdToConnection map[uint64]net.Conn

type Config struct{}

type Repository interface {
	CreateMatch(ctx context.Context, game entity.Game) (string, error)
}

type Service struct {
	config      Config
	repository  Repository
	webSocket   websocket.WebSocket
	connections IdToConnection
}

func NewService(config Config, repo Repository, ws websocket.WebSocket) Service {
	return Service{
		config:      config,
		repository:  repo,
		webSocket:   ws,
		connections: IdToConnection{},
	}
}

func (svc Service) ProcessGame(ctx echo.Context, request ProcessGameRequest) (ProcessGameResponse, error) {
	const op = "game.processGame"

	connection, _, _, err := svc.webSocket.Upgrade(ctx.Request(), ctx.Response())
	if err != nil {
		logger.Logger.Named(op).Error("error in initializing websocket", zap.Error(err))
		return ProcessGameResponse{}, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	}

	id, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		logger.Logger.Named(op).Error("error in converting id to Uint", zap.Error(err), zap.String("id", request.Id))
		return ProcessGameResponse{}, err
	}

	svc.connections[id] = *connection
	err = svc.writeMessage([]uint64{id}, "salam")
	if err != nil {
		return ProcessGameResponse{}, err
	}
	return ProcessGameResponse{}, nil
}

func (svc Service) ConsumeMatchCreated(message []byte) error {
	fmt.Println("==========> message: ", string(message))
	return nil
}

func (svc Service) writeMessage(ids []uint64, msg string) error {
	const op = "game.service.writeMessage"

	for _, id := range ids {
		if connection, exists := svc.connections[id]; !exists {
			return richerror.New(op).WithMessage(fmt.Sprintf("id: %d not found", id))
		} else {
			err := svc.webSocket.WriteServerData(connection, websocket.OpText, msg)
			if err != nil {
				logger.Logger.Named(op).Error("Error in writing message to client", zap.Error(err))
			}
		}
	}
	return nil
}
