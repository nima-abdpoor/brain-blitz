package claim

import (
	"BrainBlitz.com/game/internal/core/service"
	middleware "BrainBlitz.com/game/internal/middleware/constants"
	"BrainBlitz.com/game/pkg/richerror"
	"github.com/gin-gonic/gin"
)

func GetClaimsFromEchoContext(c *gin.Context) (service.Claim, error) {
	const op = "claim.GetClaimsFromEchoContext"
	if result, exists := c.Get(middleware.UserId); exists {
		if claim, possible := result.(service.Claim); possible {
			return claim, nil
		}
	}
	return service.Claim{}, richerror.New(op).WithKind(richerror.KindUnexpected)
}
