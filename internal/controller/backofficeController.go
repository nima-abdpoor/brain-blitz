package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (uc HttpController) InitBackofficeController(api *gin.RouterGroup) {
	api = api.Group("/backoffice")
	api.GET("/listUsers", uc.ListUsers)
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
