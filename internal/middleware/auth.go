package middleware

import (
	"BrainBlitz.com/game/internal/core/port/service"
	auth "BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/errmsg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

func Auth(authService service.AuthGenerator) echo.MiddlewareFunc {
	const op = "middleware.Auth"
	var id string
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			id = ctx.Param("id")
			if len(id) == 0 || id == "0" {
				return ctx.JSON(http.StatusBadRequest, "Invalid Id")
			}
			token := ctx.Request().Header.Get("Authorization")
			if len(token) == 0 {
				return ctx.JSON(http.StatusForbidden, errmsg.InvalidAuthentication)
				//ctx.Abort()
				//return
			}
			valid, data, err := authService.ValidateToken([]string{"user", "role"}, token)
			if err != nil {
				logger.Logger.Named(op).Error("error in Validating token", zap.Error(err))
				return ctx.JSON(http.StatusForbidden, errmsg.InvalidAuthentication)
			}
			userId, possible := data["user"].(string)
			if !possible {
				logger.Logger.Named(op).Error("cant cast data of data[user]")
				return ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
			}
			role, possible := data["role"].(string)
			if !possible {
				//todo add metrics
				logger.Logger.Named(op).Error("cant cast data of data[role]")
				return ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
			}
			if !valid || userId != id {
				// todo add metrics
				return ctx.JSON(http.StatusForbidden, errmsg.AccessDenied)
			}
			ctx.Set(middleware.UserId, auth.Claim{
				UserId: id,
				Role:   role,
			})
			return next(ctx)
		}
	}
}
