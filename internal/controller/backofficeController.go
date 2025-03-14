package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (uc HttpController) InitBackofficeController(api *echo.Group) {
	api = api.Group("/backoffice")
	api.GET("/:id/listUsers", uc.ListUsers)
}

func (uc HttpController) ListUsers(ctx echo.Context) error {
	var req request.ListUserRequest
	code := http.StatusOK
	res, err := uc.Service.BackofficeUserService.ListUsers(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	} else {
		return ctx.JSON(code, res)
	}
}
