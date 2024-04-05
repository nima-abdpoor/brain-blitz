package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
)

type MatchMakingService interface {
	AddToWaitingList(request *request.AddToWaitingListRequest) (response.AddToWaitingListResponse, error)
}
