package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (svc Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"message": "everything is good!",
	})
}
