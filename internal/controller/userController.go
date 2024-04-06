package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

func (uc HttpController) InitUserController(api *gin.RouterGroup) {
	api.POST("/signUp", uc.SignUp)
	api.GET("/signIn", uc.SignIn)
	api.GET("/:id/profile", middleware.Auth(uc.Service.AuthService), uc.Profile)
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
	ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
	if err != nil {
		log.Println("couldn't cast to Claim")
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		ctx.Abort()
		return
	}
	id, err := strconv.ParseInt(ctxClaim.UserId, 10, 64)
	resp, err := uc.Service.UserService.Profile(id)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
	} else {
		code = http.StatusOK
		ctx.JSON(code, resp)
	}
}
