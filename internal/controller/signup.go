package controller

import (
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (uc HttpController) SignUp(ctx *gin.Context) {
	req, err := uc.parseRequest(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &invalidRequestResponse)
		return
	}
	resp := uc.Service.UserService.SignUp(req)
	ctx.JSON(http.StatusOK, resp)
}

func (uc HttpController) parseRequest(ctx *gin.Context) (*request.SignUpRequest, error) {
	var req request.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

var (
	invalidRequestResponse = &response.Response{
		ErrorCode:    error_code.BadRequest,
		ErrorMessage: error_code.InvalidRequestErrMsg,
		Status:       false,
	}
)
