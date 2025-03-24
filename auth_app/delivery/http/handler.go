package http

import (
	"BrainBlitz.com/game/auth_app/service"
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/logger"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	service service.Service
	logger  logger.SlogAdapter
}

func NewHandler(service service.Service, logger logger.SlogAdapter) Handler {
	return Handler{
		service: service,
		logger:  logger,
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
		msg, code := errApp.ToHTTPJson(err)
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
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h Handler) ValidateToken(ctx echo.Context) error {
	var request service.ValidateTokenRequest
	if len(ctx.Request().Header.Get(echo.HeaderAuthorization)) == 0 {
		return ctx.JSON(http.StatusUnauthorized, service.ValidateTokenResponse{
			Valid: false,
		})
	}
	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, service.ValidateTokenResponse{
			Valid: false,
		})
	}

	request.Token = ctx.Request().Header.Get(echo.HeaderAuthorization)

	response, err := h.service.ValidateToken(ctx.Request().Context(), request)
	if err != nil {
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
	}

	if response.Valid {
		for _, data := range response.AdditionalData {
			switch data.Key {
			case "id":
				ctx.Response().Header().Set("X-User-ID", data.Value)
			case "role":
				ctx.Response().Header().Set("X-User-Role", data.Value)
			}
		}
		if result, err := json.Marshal(response.AdditionalData); err == nil {
			ctx.Response().Header().Set("X-Auth-Data", string(result))
		} else {
			h.logger.Error("marshal response error", "error", err.Error())
		}
	}

	return ctx.JSON(http.StatusOK, response)
}
