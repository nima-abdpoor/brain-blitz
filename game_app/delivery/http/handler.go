package http

import (
	"BrainBlitz.com/game/game_app/service"
	"BrainBlitz.com/game/pkg/httpmsg"
	"BrainBlitz.com/game/pkg/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) Handler {
	return Handler{
		service: service,
	}
}

func (h Handler) ProcessGame(ctx echo.Context) error {
	const op = "game.init"
	response, err := h.service.ProcessGame(ctx, service.ProcessGameRequest{
		Id: ctx.Param("id"),
	})

	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, response)
}
