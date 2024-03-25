package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	mwConstants "BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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
	log.Println("adsfasdfasdfdsf")
	if userId, exists := ctx.Get(mwConstants.UserId); exists {
		id, possible := userId.(int64)
		if !possible {
			ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
		}
		log.Println(id)
		resp, err := uc.Service.UserService.Profile(id)
		if err != nil {
			msg, code := httpmsg.Error(err)
			ctx.JSON(code, msg)
		} else {
			code = http.StatusOK
			ctx.JSON(code, resp)
		}
	} else {
		ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
		return
	}
}
