package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (uc HttpController) InitNotificationController(api *echo.Group) {
	api = api.Group("/game")
	api.GET("/:id/init", uc.initGame,
		middleware.Auth(uc.Service.AuthService),
		middleware.Presence(uc.Service.Presence),
	)
}

func (uc HttpController) initGame(ctx echo.Context) error {
	const op = "notification.initGame"

	_, err := uc.Service.Notification.InitGame(ctx, &request.InitGameRequest{
		Id: ctx.Param("id"),
	})
	if err != nil {
		// todo add metrics
		logger.Logger.Named(op).Error("error in initializing game notification", zap.Error(err))
	}
	return err
}
