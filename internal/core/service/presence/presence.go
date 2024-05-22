package presenceservice

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	s "BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/richerror"
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
		return response.UpsertPresenceResponse{}, richerror.New(op).WithError(err)
	}
	return response.UpsertPresenceResponse{}, nil
}
