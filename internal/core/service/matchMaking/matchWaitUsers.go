package matchMakingHandler

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
)

func (s Service) MatchWaitUsers(ctx context.Context, req *request.MatchWaitedUsersRequest) (response.MatchWaitedUsersResponse, error) {
	const op = "matchMakingHandler.MatchWaitUsers"
	var rErr error = nil
	for _, category := range entity.GetCategories() {
		result, err := s.repo.GetWaitingListByCategory(ctx, category)
		for _, res := range result {
			fmt.Println(op, res)
		}
		if err != nil {
			rErr = richerror.New(op).WithError(err)
		}
	}
	return response.MatchWaitedUsersResponse{}, rErr
}
