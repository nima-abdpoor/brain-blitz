package controller

import (
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/gin-gonic/gin"
)

type HttpController struct {
	Gin     *gin.Engine
	Service service.Service
}

func NewController(gin *gin.Engine, us service.Service) HttpController {
	return HttpController{
		Gin:     gin,
		Service: us,
	}
}

func (uc HttpController) InitRouter() {
	api := uc.Gin.Group("/api/v1")
	uc.InitUserController(api)
}
