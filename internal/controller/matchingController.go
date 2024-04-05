package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	mwConstants "BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (uc HttpController) InitMatchingController(api *gin.RouterGroup) {
	api = api.Group("/matching")
	api.POST("/:id/addToWaitingList",
		middleware.Auth(uc.Service.AuthService),
		uc.addToWaitingList,
	)
}

func (uc HttpController) addToWaitingList(ctx *gin.Context) {
	var req request.AddToWaitingListRequest
	code := http.StatusOK
	if userId, exists := ctx.Get(mwConstants.UserId); exists {
		id, possible := userId.(int64)
		if !possible {
			ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
		}
		log.Println(id)
		err := ctx.BindJSON(&req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
		}
		req.UserId = id
		fmt.Println(req)
		res, err := uc.Service.MatchMakingService.AddToWaitingList(&req)
		if err != nil {
			msg, code := httpmsg.Error(err)
			ctx.JSON(code, msg)
		} else {
			ctx.JSON(code, res)
		}
	} else {
		log.Println("could not get userId from context!")
		ctx.JSON(http.StatusInternalServerError, errmsg.SomeThingWentWrong)
		return
	}
}
