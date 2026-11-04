package http

import (
	"context"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

// Identity is supplied by the gateway's ForwardAuth, never a query parameter.
func (h Handler) Statistics(c echo.Context) error {
	id, err := strconv.ParseUint(c.Request().Header.Get("X-User-ID"), 10, 64)
	if err != nil || id == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Authentication required"})
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 4*time.Second)
	defer cancel()
	stats, err := h.service.PlayerStatistics(ctx, id)
	if err != nil {
		h.logger.Error("statistics unavailable", "error", err.Error())
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"message": "Statistics temporarily unavailable"})
	}
	return c.JSON(http.StatusOK, stats)
}
