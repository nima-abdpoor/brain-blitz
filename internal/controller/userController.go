package controller

import (
	"BrainBlitz.com/game/internal/core/common/router"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (uc HttpController) InitUserController(api *gin.RouterGroup) {
	api.Group(api.BasePath() + "/user")
	router.Post(api, "/signup", uc.SignUp)
	router.Get(api, "/signin", uc.SignIn)
	router.Get(api, "/:id/profile", uc.Profile)
}

func (uc HttpController) SignIn(ctx *gin.Context) {
	var req request.SignInRequest
	code := http.StatusOK
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errmsg.InvalidUserNameOrPasswordErrMsg)
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
	code := http.StatusOK
	var req request.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errmsg.InvalidUserNameOrPasswordErrMsg)
		return
	}
	resp, err := uc.Service.UserService.SignUp(&req)
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
