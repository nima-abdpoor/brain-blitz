package http

import (
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/services/match_app/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	Service service.Service
	Logger  logger.Logger
}

func NewHandler(userService service.Service, logger logger.Logger) Handler {
	return Handler{
		Service: userService,
		Logger:  logger,
	}
}

func (handler Handler) addToWaitingList(ctx echo.Context) error {
	const op = "controller.addToWaitingList"

	var req service.AddToWaitingListRequest
	code := http.StatusOK

	id := ctx.Request().Header.Get("X-User-ID")
	if len(id) < 1 {
		handler.Logger.Info(op, "message", "Invalid X-User-ID", "id", id)
		return ctx.JSON(http.StatusBadRequest, errmsg.MessageMissingXUserId)
	}
	req.UserId = id

	err := ctx.Bind(&req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
	}
	res, err := handler.Service.AddToWaitingList(ctx.Request().Context(), req)
	if err != nil {
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
	} else {
		return ctx.JSON(code, res)
	}
}
