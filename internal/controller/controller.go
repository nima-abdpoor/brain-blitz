package controller

import (
	"BrainBlitz.com/game/internal/core/common/router"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	invalidRequestResponse = &response.Response{
		ErrorCode:    error_code.InvalidRequest,
		ErrorMessage: error_code.InvalidRequestErrMsg,
		Status:       false,
	}
)

type UserController struct {
	gin         *gin.Engine
	userService service.UserService
}

func NewUserController(gin *gin.Engine, us service.UserService) UserController {
	return UserController{
		gin:         gin,
		userService: us,
	}
}

func (uc UserController) InitRouter() {
	api := uc.gin.Group("/api/v1")
	router.Post(api, "/signup", uc.signUp)
}

func (uc UserController) signUp(ctx *gin.Context) {
	req, err := uc.parseRequest(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusOK, &invalidRequestResponse)
		return
	}
	resp := uc.userService.SignUp(req)
	ctx.JSON(http.StatusOK, resp)
}

func (uc UserController) parseRequest(ctx *gin.Context) (*request.SignUpRequest, error) {
	var req request.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}
