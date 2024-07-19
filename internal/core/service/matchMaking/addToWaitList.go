package matchMakingHandler

import (
	"BrainBlitz.com/game/adapter/broker"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"log"
	"time"
)

type Service struct {
	repo            repository.MatchMakingRepository
	presenceClient  repository.PresenceClient
	publisherBroker broker.PublisherBroker
	consumerBroker  broker.ConsumerBroker
	config          Config
}

type Config struct {
	WaitingTimeout time.Duration `koanf:"waiting_timeout"`
	LeastPresence  time.Duration `koanf:"least_presence"`
}

func NewMatchMakingService(repo repository.MatchMakingRepository, presenceClient repository.PresenceClient, publisherBroker broker.PublisherBroker, consumerBroker broker.ConsumerBroker, config Config) service.MatchMakingService {
	return Service{
		repo:            repo,
		presenceClient:  presenceClient,
		publisherBroker: publisherBroker,
		consumerBroker:  consumerBroker,
		config:          config,
	}
}

func (s Service) AddToWaitingList(request *request.AddToWaitingListRequest) (response.AddToWaitingListResponse, error) {
	const op = "matchMakingHandler.AddToWaitingList"

	err := s.repo.AddToWaitingList(entity.MapToCategory(request.Category), request.UserId)
	if err != nil {
		log.Printf("error in %s:%v", op, err)
		return response.AddToWaitingListResponse{},
			richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err).WithMessage(errmsg.SomeThingWentWrong)
	}
	resp := response.AddToWaitingListResponse{
		Timeout: s.config.WaitingTimeout,
	}

	return resp, nil
}
