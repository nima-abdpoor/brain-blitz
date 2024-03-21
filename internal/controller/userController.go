package controller

import (
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

func (uc HttpController) SignUp(ctx *gin.Context) {
	req, err := uc.parseRequest(ctx)
	code := http.StatusOK
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &invalidRequestResponse)
		return
	}
	resp, err := uc.Service.UserService.SignUp(req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
	} else {
		ctx.JSON(code, resp)
	}
}

func (uc HttpController) Profile(ctx *gin.Context) {
	code := http.StatusBadRequest
	var profileReq request.ProfileRequest
	if err := ctx.ShouldBindUri(&profileReq); err != nil {
		ctx.JSON(code, "Invalid Id")
		return
	}
	token := ctx.Request.Header.Get("Authorization")
	if len(token) == 0 {
		ctx.JSON(http.StatusForbidden, "authentication required")
		return
	}
	resp, err := uc.Service.UserService.Profile(strconv.FormatInt(profileReq.ID, 10), token)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
	} else {
		code = http.StatusOK
		ctx.JSON(code, resp)
	}
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
