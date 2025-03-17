package middleware

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/labstack/echo/v4"
)

func Presence(service service.PresenceService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = "middleware.presence"

			if _, err := service.Upsert(c.Request().Context(), request.UpsertPresenceRequest{
				UserID: "",
			}); err != nil {
				//logger.Logger.Named(op).Error("error In upsetting", zap.String("ctxClaim.UserId", ""), zap.Error(err))
			}
			return next(c)
		}
	}
}
