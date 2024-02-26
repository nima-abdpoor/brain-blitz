package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (uc HttpController) SignIn(ctx *gin.Context) {
	var req request.SingInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &invalidRequestResponse)
		return
	}
	res := uc.Service.UserService.SignIn(&req)
	ctx.JSON(res.ErrorCode, res)
}
