package controller

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/middleware"
	"BrainBlitz.com/game/pkg/claim"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func (uc HttpController) InitMatchingController(api *echo.Group) {
	api = api.Group("/matching")
	api.POST("/:id/addToWaitingList",
		uc.addToWaitingList,
		middleware.Auth(uc.Service.AuthService),
		middleware.Presence(uc.Service.Presence),
	)
}

func (uc HttpController) addToWaitingList(ctx echo.Context) error {
	var req request.AddToWaitingListRequest
	code := http.StatusOK
	ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
	if err != nil {
		log.Println("couldn't cast to Claim")
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}
	err = ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
		return nil
	}
	req.UserId = ctxClaim.UserId
	res, err := uc.Service.MatchMakingService.AddToWaitingList(&req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		return nil
	} else {
		ctx.JSON(code, res)
		return nil
	}
	return nil
}
