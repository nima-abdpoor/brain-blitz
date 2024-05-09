package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"time"
)

func TimeoutMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		timeout := 1 * time.Second
		ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
		defer cancel()
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
