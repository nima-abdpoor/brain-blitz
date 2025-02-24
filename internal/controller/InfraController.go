package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (uc HttpController) InitInfraController(api *echo.Group) {
	api.GET("/metrics", uc.Metrics)
}

func (uc HttpController) Metrics(ctx echo.Context) error {
	promhttp.Handler()
	return nil
}
