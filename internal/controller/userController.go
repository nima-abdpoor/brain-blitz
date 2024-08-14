package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func (uc HttpController) InitUserController(api *echo.Group) {
	api.POST("/signUp", uc.SignUp)
	api.GET("/signIn", uc.SignIn)
	api.GET("/:id/profile", uc.Profile,
		middleware.Auth(uc.Service.AuthService),
		middleware.Presence(uc.Service.Presence))
}

func (uc HttpController) SignIn(ctx echo.Context) error {
	var req request.SignInRequest
	code := http.StatusOK
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.InvalidUserNameOrPasswordErrMsg)
	}
	res, err := uc.Service.UserService.SignIn(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	} else {
		return ctx.JSON(code, res)
	}
}

func (uc HttpController) SignUp(ctx echo.Context) error {
	code := http.StatusOK
	var req request.SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.InvalidUserNameOrPasswordErrMsg)
	}
	resp, err := uc.Service.UserService.SignUp(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	} else {
		return ctx.JSON(code, resp)
	}
}

func (uc HttpController) Profile(ctx echo.Context) error {
	const op = "controller.Profile"
	code := http.StatusBadRequest
	ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
	if err != nil {
		//todo add to metrics
		logger.Logger.Named(op).Error("couldn't cast to Claim claim.GetClaimsFromEchoContext")
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}
	id, err := strconv.ParseInt(ctxClaim.UserId, 10, 64)
	resp, err := uc.Service.UserService.Profile(id)
	if err != nil {
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	} else {
		code = http.StatusOK
		return ctx.JSON(code, resp)
	}
}
