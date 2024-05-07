package matchMakingHandler

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"fmt"
	"time"
)

func (s Service) MatchWaitUsers(req *request.MatchWaitedUsersRequest) (response.MatchWaitedUsersResponse, error) {
	fmt.Println("waited users matched.", time.Now())
	return response.MatchWaitedUsersResponse{}, nil
}
