package matchMakingHandler

import (
	"BrainBlitz.com/game/adapter/broker"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/logger"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/richerror"
	"go.uber.org/zap"
	"time"
)

type Service struct {
	repo            repository.MatchMakingRepository
	presenceClient  repository.PresenceClient
	publisherBroker broker.PublisherBroker
	config          Config
}

type Config struct {
	WaitingTimeout time.Duration `koanf:"waiting_timeout"`
	LeastPresence  time.Duration `koanf:"least_presence"`
}

func NewMatchMakingService(repo repository.MatchMakingRepository, presenceClient repository.PresenceClient, publisherBroker broker.PublisherBroker, config Config) service.MatchMakingService {
	return Service{
		repo:            repo,
		presenceClient:  presenceClient,
		publisherBroker: publisherBroker,
		config:          config,
	}
}

func (s Service) AddToWaitingList(request *request.AddToWaitingListRequest) (response.AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := s.repo.AddToWaitingList(entity.MapToCategory(request.Category), request.UserId)
	if err != nil {
		logger.Logger.Named(op).Error("add to waiting list failed", zap.String("request.UserId", request.UserId), zap.Error(err))
		return response.AddToWaitingListResponse{},
			richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err).WithMessage(errmsg.SomeThingWentWrong)
	}
	resp := response.AddToWaitingListResponse{
		Timeout: s.config.WaitingTimeout,
	}

	return resp, nil
}
