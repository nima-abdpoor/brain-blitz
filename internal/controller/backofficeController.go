package controller

import (
	entity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (uc HttpController) InitBackofficeController(api *gin.RouterGroup) {
	api = api.Group("/backoffice")
	api.GET("/:id/listUsers",
		middleware.Auth(uc.Service.AuthService),
		middleware.AccessCheck(uc.Service.AuthorizationService, entity.UserListPermission, entity.UserDeletePermission),
		uc.ListUsers)
}

func (uc HttpController) ListUsers(ctx *gin.Context) {
	var req request.ListUserRequest
	code := http.StatusOK
	res, err := uc.Service.BackofficeUserService.ListUsers(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
	} else {
		ctx.JSON(code, res)
	}
}
