package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	entity "BrainBlitz.com/game/entity/game"
	"context"
	"github.com/labstack/echo/v4"
)

type Config struct{}

type Repository interface {
	CreateMatch(ctx context.Context, game entity.Game) (string, error)
}

type Service struct {
	config     Config
	repository Repository
	webSocket  websocket.WebSocket
}

func NewService(config Config, repo Repository) Service {
	return Service{
		config:     config,
		repository: repo,
	}
}

func (svc Service) ProcessGame(ctx echo.Context, request ProcessGameRequest) error {
	const op = "game.processGame"

	connection, _, _, err := svc.webSocket.Upgrade(ctx.Request(), ctx.Response())
}
