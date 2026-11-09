package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) GameOptions(c echo.Context) error {
	return c.JSON(http.StatusOK, h.service.GameOptions())
}
