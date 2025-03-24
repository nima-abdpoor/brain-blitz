package presenceservice

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	s "BrainBlitz.com/game/internal/core/port/service"
	"context"
	"fmt"
	"time"
)

type Config struct {
	Prefix         string        `koanf:"prefix"`
	ExpirationTime time.Duration `koanf:"expiration_time"`
}

type service struct {
	repo   repository.PresenceRepository
	config Config
}

type GetPresenceService struct {
	getPresenceRepo repository.PresenceClient
}

func New(repo repository.PresenceRepository, config Config) s.PresenceService {
	return service{
		repo:   repo,
		config: config,
	}
}

func (s service) Upsert(context context.Context, request request.UpsertPresenceRequest) (response.UpsertPresenceResponse, error) {
	const op = "presenceservice.Upsert"
	if err :=
		s.repo.Upsert(
			context,
			fmt.Sprintf("%s:%s", s.config.Prefix, request.UserID),
			time.Now().UnixMilli(),
			s.config.ExpirationTime); err != nil {
		//return response.UpsertPresenceResponse{}, richerror.New(op).WithError(err)
	}
	return response.UpsertPresenceResponse{}, nil
}

func (s GetPresenceService) GetPresence(ctx context.Context, request request.GetPresenceRequest) (response.GetPresenceResponse, error) {
	const op = "presenceservice.GetPresenceByID"

	if rsp, err := s.getPresenceRepo.GetPresence(ctx, request.UserID); err != nil {
		return response.GetPresenceResponse{}, err
		//return response.GetPresenceResponse{}, richerror.New(op).WithError(err)
	} else {
		return response.GetPresenceResponse{
			UserIdToTimestamp: rsp,
		}, nil
	}
}
