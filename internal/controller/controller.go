package controller

import (
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type HttpController struct {
	Echo    *echo.Echo
	Service service.Service
}

type InfraHttpController struct {
	Handler http.Handler
}

func NewController(echo *echo.Echo, us service.Service) HttpController {
	return HttpController{
		Echo:    echo,
		Service: us,
	}
}

func NewInfraHttpController(handler http.Handler) InfraHttpController {
	return InfraHttpController{
		Handler: handler,
	}
}

func (uc HttpController) InitRouter() {
	api := uc.Echo.Group("/api/v1")
	uc.InitBackofficeController(api)
}

func (uc InfraHttpController) InitRouter() {
	http.Handle("/metrics", promhttp.Handler())
}
