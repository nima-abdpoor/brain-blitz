package claim

import (
	"BrainBlitz.com/game/internal/core/service"
	middleware "BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/pkg/richerror"
	"github.com/labstack/echo/v4"
)

func GetClaimsFromEchoContext(c echo.Context) (service.Claim, error) {
	const op = "claim.GetClaimsFromEchoContext"
	result := c.Get(middleware.UserId)
	if result == nil {
		return service.Claim{}, richerror.New(op).WithKind(richerror.KindUnexpected)
	}
	if claim, possible := result.(service.Claim); possible {
		return claim, nil
	}
	return service.Claim{}, richerror.New(op).WithKind(richerror.KindUnexpected)
}
