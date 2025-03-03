package middleware

import (
	entity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/metrics"
	"BrainBlitz.com/game/pkg/claim"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"net/http"
)

func AccessCheck(authService service.AuthorizationService, permissions ...entity.PermissionTitle) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
			if err != nil {
				metrics.FailedClaimCounter.Inc()
				msg, code := httpmsg.Error(err)
				return ctx.JSON(code, msg)
			}
			hasAccess, err := authService.HasAccess(entity.MapToRoleEntity(ctxClaim.Role), permissions...)
			if err != nil {
				msg, code := httpmsg.Error(err)
				return ctx.JSON(code, msg)
			}
			if !hasAccess {
				return ctx.JSON(http.StatusForbidden, errmsg.PermissionRequired)
			}
			return next(ctx)
		}
	}
}
