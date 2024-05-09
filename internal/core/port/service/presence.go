package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"context"
)

type PresenceService interface {
	Upsert(context context.Context, request request.UpsertPresenceRequest) (response.UpsertPresenceResponse, error)
}
