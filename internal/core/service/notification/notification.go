package notification

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
)

type Service struct {
	config Config
}

type Config struct {
}

func New(config Config) service.Notification {
	return Service{
		config: config,
	}
}

func (s Service) InitGame(ctx echo.Context, req *request.InitGameRequest) (request.InitGameResponse, error) {
	const op = "notification.InitGame"

	connection, _, _, err := ws.UpgradeHTTP(ctx.Request(), ctx.Response())
	if err != nil {
		logger.Logger.Named(op).Error("error in initializing websocket", zap.Error(err))
		return request.InitGameResponse{}, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
	}

	return request.InitGameResponse{}, nil
}

func readMessage(rw io.ReadWriter) (message string, opCode int, err error) {
	for {
		msg, opCode, err := wsutil.ReadClientData(rw)
		if err != nil {
			logger.Logger.Named()
			fmt.Println("error in reading message...", err)
		}

		fmt.Println("message:", string(msg), "opCODE:", opCode)
	}
}
