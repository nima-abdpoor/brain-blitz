package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"github.com/labstack/echo/v4"
)

type Notification interface {
	InitGame(ctx echo.Context, request *request.InitGameRequest) (request.InitGameResponse, error)
}
