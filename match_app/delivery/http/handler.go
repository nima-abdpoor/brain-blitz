package http

import (
	"BrainBlitz.com/game/match_app/service"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	Service service.Service
}

func NewHandler(userService service.Service) Handler {
	return Handler{
		Service: userService,
	}
}

func (handler Handler) addToWaitingList(ctx echo.Context) error {
	const op = "controller.addToWaitingList"

	var req service.AddToWaitingListRequest
	code := http.StatusOK

	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
		return nil
	}
	res, err := handler.Service.AddToWaitingList(ctx.Request().Context(), req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		return nil
	} else {
		ctx.JSON(code, res)
		return nil
	}
}
