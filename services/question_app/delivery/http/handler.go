package http

import (
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/services/question_app/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	Service service.Service
	Logger  logger.Logger
}

func NewHandler(service service.Service, logger logger.Logger) Handler {
	return Handler{
		Service: service,
		Logger:  logger,
	}
}

func (h Handler) AddQuestion(ctx echo.Context) error {
	var req service.AddQuestionRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidInputErrMsg})
	}

	_, err := h.Service.AddQuestion(ctx.Request().Context(), req)
	if err != nil {
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, service.AddQuestionResponse{})
}
