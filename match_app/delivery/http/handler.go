package http

import (
	auth "BrainBlitz.com/game/internal/core/service"
	middleware "BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/match_app/service"
	"BrainBlitz.com/game/metrics"
	"BrainBlitz.com/game/pkg/claim"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/httpmsg"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	Service service.Service
}

func NewHandler(userService service.Service) Handler {
	return Handler{
		Service: userService,
	}
}

func (handler Handler) addToWaitingList(ctx echo.Context) error {
	const op = "controller.addToWaitingList"
	//todo we should remove this
	id := ctx.Param("id")
	ctx.Set(middleware.UserId, auth.Claim{
		UserId: id,
	})

	var req service.AddToWaitingListRequest
	code := http.StatusOK
	ctxClaim, err := claim.GetClaimsFromEchoContext(ctx)
	if err != nil {
		metrics.FailedClaimCounter.Inc()
		logger.Logger.Named(op).Error("couldn't cast to Claim claim.GetClaimsFromEchoContext")
		msg, code := httpmsg.Error(err)
		return ctx.JSON(code, msg)
	}
	err = ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errmsg.InvalidCategory)
		return nil
	}
	req.UserId = ctxClaim.UserId
	res, err := handler.Service.AddToWaitingList(ctx.Request().Context(), req)
	if err != nil {
		msg, code := httpmsg.Error(err)
		ctx.JSON(code, msg)
		return nil
	} else {
		ctx.JSON(code, res)
		return nil
	}
}
