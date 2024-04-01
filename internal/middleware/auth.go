package middleware

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	auth "BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/pkg/errmsg"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

func Auth(authService service.AuthGenerator) gin.HandlerFunc {
	var profileReq request.ProfileRequest
	return func(ctx *gin.Context) {
		if err := ctx.ShouldBindUri(&profileReq); err != nil {
			ctx.JSON(http.StatusBadRequest, "Invalid Id")
			ctx.Abort()
			return
		}
		token := ctx.Request.Header.Get("Authorization")
		if len(token) == 0 {
			ctx.JSON(http.StatusForbidden, errmsg.InvalidAuthentication)
			ctx.Abort()
			return
		}
		valid, data, err := authService.ValidateToken([]string{"user", "role"}, token)
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusForbidden, errmsg.InvalidAuthentication)
			ctx.Abort()
			return
		}
		userId, possible := data["user"].(string)
		if !possible {
			log.Println("cant cast data")
			ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
			ctx.Abort()
			return
		}
		role, possible := data["role"].(string)
		if !possible {
			log.Println("cant cast data")
			ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
			ctx.Abort()
			return
		}
		if !valid || userId != strconv.FormatInt(profileReq.ID, 10) {
			log.Println("is not valid", valid, data["user"], profileReq.ID)
			ctx.JSON(http.StatusForbidden, errmsg.AccessDenied)
			ctx.Abort()
			return
		}
		ctx.Set(middleware.UserId, auth.Claim{
			UserId: strconv.FormatInt(profileReq.ID, 10),
			Role:   role,
		})
		ctx.Next()
	}
}
