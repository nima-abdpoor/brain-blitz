package middleware

import (
	entity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AccessCheck(authService service.AuthorizationService, permissions ...entity.PermissionTitle) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
		if err != nil {
			msg, code := httpmsg.Error(err)
			ctx.JSON(code, msg)
			ctx.Abort()
			return
		}
		hasAccess, err := authService.HasAccess(entity.MapToRoleEntity(ctxClaim.Role), permissions...)
		if err != nil {
			msg, code := httpmsg.Error(err)
			ctx.JSON(code, msg)
			ctx.Abort()
			return
		}
		if !hasAccess {
			ctx.JSON(http.StatusForbidden, errmsg.PermissionRequired)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
