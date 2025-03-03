package http

import (
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/statuscode"
	"BrainBlitz.com/game/pkg/validator"
	"BrainBlitz.com/game/user_app/service"
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

func (h Handler) SignUp(ctx echo.Context) error {
	var req service.SignUpRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidUserNameOrPasswordErrMsg})
	}

	res, err := h.Service.SignUp(ctx.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		return ctx.JSON(statuscode.MapToHTTPStatusCode(err.(errmsg.ErrorResponse)), err)
	}

	return ctx.JSON(http.StatusOK, service.SignUpResponse{
		DisplayName: res.DisplayName,
	})
}

func (h Handler) Login(ctx echo.Context) error {
	var req service.LoginRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidUserNameOrPasswordErrMsg})
	}

	res, err := h.Service.Login(ctx.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		return ctx.JSON(statuscode.MapToHTTPStatusCode(err.(errmsg.ErrorResponse)), err)
	}

	return ctx.JSON(http.StatusOK, service.LoginResponse{
		ID:           res.ID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

func (h Handler) Profile(ctx echo.Context) error {
	var req service.ProfileRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.AccessDenied})
	}

	res, err := h.Service.Profile(ctx.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return ctx.JSON(vErr.StatusCode(), vErr)
		}
		return ctx.JSON(statuscode.MapToHTTPStatusCode(err.(errmsg.ErrorResponse)), err)
	}

	return ctx.JSON(http.StatusOK, service.ProfileResponse{
		ID:          res.ID,
		Username:    res.Username,
		DisplayName: res.DisplayName,
		Role:        res.Role,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
	})
}
