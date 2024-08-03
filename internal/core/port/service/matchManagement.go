package service

import "BrainBlitz.com/game/internal/core/model/request"

type MatchManagementService interface {
	StartMatchCreator(req request.StartMatchCreatorRequest) (request.StartMatchCreatorRequest, error)
}
