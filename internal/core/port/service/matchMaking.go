package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"context"
)

type MatchMakingService interface {
	AddToWaitingList(request *request.AddToWaitingListRequest) (response.AddToWaitingListResponse, error)
	MatchWaitUsers(ctx context.Context, request *request.MatchWaitedUsersRequest) (response.MatchWaitedUsersResponse, error)
	StartMatchMaker()
}
