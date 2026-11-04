package http

import (
	"BrainBlitz.com/game/pkg/timeoptions"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) WithTimeOptions(store timeoptions.Store) Handler { h.timeOptions = &store; return h }
func (h Handler) TimeOptions(c echo.Context) error                { return h.listTimeOptions(c, false) }
func (h Handler) AdminTimeOptions(c echo.Context) error {
	if c.Request().Header.Get("X-User-Role") != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{"message": "Administrator access required"})
	}
	return h.listTimeOptions(c, true)
}
func (h Handler) listTimeOptions(c echo.Context, all bool) error {
	if h.timeOptions == nil {
		return c.JSON(503, map[string]string{"message": "Time options unavailable"})
	}
	options, err := h.timeOptions.List(c.Request().Context(), all)
	if err != nil {
		h.Logger.Error("time options failed", "error", err)
		return c.JSON(503, map[string]string{"message": "Time options unavailable"})
	}
	return c.JSON(200, options)
}
func (h Handler) PutTimeOption(c echo.Context) error {
	if c.Request().Header.Get("X-User-Role") != "admin" {
		return c.JSON(403, map[string]string{"message": "Administrator access required"})
	}
	var option timeoptions.Option
	if err := c.Bind(&option); err != nil {
		return c.JSON(400, map[string]string{"message": "Invalid time option"})
	}
	option.ID = c.Param("id")
	if err := option.Validate(); err != nil {
		return c.JSON(400, map[string]string{"message": err.Error()})
	}
	if h.timeOptions == nil {
		return c.JSON(503, map[string]string{"message": "Time options unavailable"})
	}
	if err := h.timeOptions.Put(c.Request().Context(), option); err != nil {
		h.Logger.Error("saving time option failed", "error", err)
		return c.JSON(503, map[string]string{"message": "Could not save time option"})
	}
	return c.JSON(200, option)
}
