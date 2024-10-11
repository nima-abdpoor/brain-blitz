package controller

import (
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/labstack/echo/v4"
)

type HttpController struct {
	Echo    *echo.Echo
	Service service.Service
}

func NewController(echo *echo.Echo, us service.Service) HttpController {
	return HttpController{
		Echo:    echo,
		Service: us,
	}
}

func (uc HttpController) InitRouter() {
	api := uc.Echo.Group("/api/v1")
	uc.InitUserController(api)
	uc.InitBackofficeController(api)
	uc.InitMatchingController(api)
	uc.InitNotificationController(api)
}
