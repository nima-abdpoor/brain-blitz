package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (uc HttpController) SignIn(ctx *gin.Context) {
	var req request.SignInRequest
	code := http.StatusOK
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &invalidRequestResponse)
		return
	}
	res, err := uc.Service.UserService.SignIn(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
	} else {
		ctx.JSON(code, res)
	}
}
