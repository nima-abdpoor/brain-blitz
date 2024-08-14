package middleware

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Presence(service service.PresenceService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = "middleware.presence"

			ctxClaim, err := claim.GetClaimsFromEchoContext(c)
			if err != nil {
				// todo add metrics
				logger.Logger.Named(op).Error("couldn't cast to Claim")
				msg, code := httpmsg.Error(err)
				return c.JSON(code, msg)
			}
			if _, err := service.Upsert(c.Request().Context(), request.UpsertPresenceRequest{
				UserID: ctxClaim.UserId,
			}); err != nil {
				logger.Logger.Named(op).Error("error In upsetting", zap.String("ctxClaim.UserId", ctxClaim.UserId), zap.Error(err))
			}
			return next(c)
		}
	}
}
