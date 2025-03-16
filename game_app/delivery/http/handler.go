package http

import (
	"BrainBlitz.com/game/game_app/service"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"BrainBlitz.com/game/pkg/validator"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type Handler struct {
	service service.Service
	logger  *slog.Logger
}

func NewHandler(service service.Service, logger *slog.Logger) Handler {
	return Handler{
		service: service,
		logger:  logger,
	}
}

func (h Handler) ProcessGame(ctx echo.Context) error {
	const op = "game.init"
	id := ctx.Request().Header.Get("X-User-ID")
	if len(id) < 1 {
		h.logger.Info("Invalid X-User-ID", "id", id)
		return ctx.JSON(http.StatusBadRequest, errmsg.MessageMissingXUserId)
	}
	response, err := h.service.ProcessGame(ctx, service.ProcessGameRequest{
		Id: id,
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
