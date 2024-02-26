package controller

import (
	"BrainBlitz.com/game/internal/core/common/router"
	"BrainBlitz.com/game/internal/core/port/service"
	"github.com/gin-gonic/gin"
)

func NewUserController(gin *gin.Engine, us service.Service) UserController {
	return UserController{
		Gin:     gin,
		Service: us,
	}
}

func (uc UserController) InitRouter() {
	api := uc.Gin.Group("/api/v1")
	router.Post(api, "/signup", uc.SignUp)
}
