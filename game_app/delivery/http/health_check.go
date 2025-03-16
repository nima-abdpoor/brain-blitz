package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s Server) healthCheck(c echo.Context) error {
	s.logger.Info(
		"Health Check",
		"id", c.Request().Header.Get("X-User-ID"),
		"role", c.Request().Header.Get("X-User-Role"),
		"data", c.Request().Header.Get("X-Auth-Data"))
	return c.JSON(http.StatusOK, echo.Map{
		"message": "everything is good!",
	})
}
