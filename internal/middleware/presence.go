package middleware

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"log"
)

func Presence(service service.PresenceService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = "middleware.presence"

			ctxClaim, err := claim.GetClaimsFromEchoContext(c)
			if err != nil {
				log.Println(op, "couldn't cast to Claim")
				msg, code := httpmsg.Error(err)
				return c.JSON(code, msg)
			}
			if _, err := service.Upsert(c.Request().Context(), request.UpsertPresenceRequest{
				UserID: ctxClaim.UserId,
			}); err != nil {
				log.Println(op, err.Error())
			}
			return next(c)
		}
	}
}
