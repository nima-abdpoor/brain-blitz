package http

import (
	"BrainBlitz.com/game/auth_app/service"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
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

func (h Handler) CreateAccessToken(ctx echo.Context) error {
	var request service.CreateAccessTokenRequest
	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidBody})
	}

	response, err := h.service.CreateAccessToken(ctx.Request().Context(), request)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h Handler) CreateRefreshToken(ctx echo.Context) error {
	var request service.CreateRefreshTokenRequest
	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidBody})
	}

	response, err := h.service.CreateRefreshToken(ctx.Request().Context(), request)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h Handler) ValidateToken(ctx echo.Context) error {
	var request service.ValidateTokenRequest
	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, nil)
	}

	response, err := h.service.ValidateToken(ctx.Request().Context(), request)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		msg, _ := httpmsg.Error(err)
		return ctx.JSON(http.StatusUnauthorized, msg)
	}

	return ctx.JSON(http.StatusOK, response)
}
