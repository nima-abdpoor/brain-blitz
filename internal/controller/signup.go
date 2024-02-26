package controller

import (
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserController struct {
	Gin     *gin.Engine
	Service service.Service
}

func (uc UserController) SignUp(ctx *gin.Context) {
	req, err := uc.parseRequest(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusOK, &invalidRequestResponse)
		return
	}
	resp := uc.Service.UserService.SignUp(req)
	ctx.JSON(http.StatusOK, resp)
}

func (uc UserController) parseRequest(ctx *gin.Context) (*request.SignUpRequest, error) {
	var req request.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

var (
	invalidRequestResponse = &response.Response{
		ErrorCode:    error_code.InvalidRequest,
		ErrorMessage: error_code.InvalidRequestErrMsg,
		Status:       false,
	}
)
