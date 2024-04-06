package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
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
	ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
	if err != nil {
		log.Println("couldn't cast to Claim")
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		ctx.Abort()
		return
	}
	err = ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
		return
	}
	req.UserId = ctxClaim.UserId
	res, err := uc.Service.MatchMakingService.AddToWaitingList(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		return
	} else {
		ctx.JSON(code, res)
		return
	}
}
