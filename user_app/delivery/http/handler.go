package http

import (
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/user_app/service"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type Handler struct {
	Service service.Service
	Logger  *slog.Logger
}

func NewHandler(userService service.Service, logger *slog.Logger) Handler {
	return Handler{
		Service: userService,
		Logger:  logger,
	}
}

func (h Handler) SignUp(ctx echo.Context) error {
	var req service.SignUpRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, errmsg.ErrorResponse{Message: errmsg.InvalidUserNameOrPasswordErrMsg})
	}

	res, err := h.Service.SignUp(ctx.Request().Context(), req)
	if err != nil {
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
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
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
	}

	return ctx.JSON(http.StatusOK, service.LoginResponse{
		ID:           res.ID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

func (h Handler) Profile(ctx echo.Context) error {
	var req service.ProfileRequest
	id := ctx.Request().Header.Get("X-User-ID")
	if len(id) < 1 {
		h.Logger.Info("Invalid X-User-ID", "id", id)
		return ctx.JSON(http.StatusBadRequest, errmsg.MessageMissingXUserId)
	}
	req.ID = id
	res, err := h.Service.Profile(ctx.Request().Context(), req)
	if err != nil {
		msg, code := errApp.ToHTTPJson(err)
		return ctx.JSON(code, msg)
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
