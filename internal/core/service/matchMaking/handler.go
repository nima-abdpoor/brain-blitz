package matchMakingHandler

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"time"
)

type Service struct {
	repo   repository.MatchMakingRepository
	config Config
}

type Config struct {
	WaitingTimeout time.Duration `koanf:"waiting_timeout"`
}

func NewMatchMakingService(repo repository.MatchMakingRepository, config Config) service.MatchMakingService {
	return Service{
		repo:   repo,
		config: config,
	}
}

func (s Service) AddToWaitingList(request *request.AddToWaitingListRequest) (response.AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := s.repo.AddToWaitingList(entity.MapToCategory(request.Category), request.UserId)
	if err != nil {
		return response.AddToWaitingListResponse{},
			richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err).WithMessage(errmsg.SomeThingWentWrong)
	}
	resp := response.AddToWaitingListResponse{
		Timeout: s.config.WaitingTimeout,
	}

	return resp, nil
}
