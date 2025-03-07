package match

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/golang/game"
	"BrainBlitz.com/game/contract/match/golang"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/match_app/service"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"
	"time"
)

type Service struct {
	repository      repository.MatchManagementRepository
	consumerBroker  broker.ConsumerBroker
	publisherBroker broker.PublisherBroker
}

func New(repo repository.MatchManagementRepository, consumer broker.ConsumerBroker, publisher broker.PublisherBroker) Service {
	return Service{
		repository:      repo,
		consumerBroker:  consumer,
		publisherBroker: publisher,
	}
}

func (s Service) StartMatchCreator(req request.StartMatchCreatorRequest) (request.StartMatchCreatorRequest, error) {
	const op = "matchMakingHandler.StartMatchMaker"
	matchMakingTopic := "matchMaking_v1_matchUsers"
	matchMakingGroup := "matchMaking"

	consumer, _ := s.consumerBroker.Consume(
		map[string]string{
			"group.id":          matchMakingGroup,
			"auto.offset.reset": "smallest",
		})
	//todo think about errors it will be logged in Consume function

	switch c := consumer.(type) {
	case *kafka.Consumer:
		{
			run := true
			defer c.Close()
			err := c.SubscribeTopics([]string{matchMakingTopic}, nil)
			if err != nil {
				logger.Logger.Named(op).Error("Error in subscribing to topics", zap.String("topic", matchMakingTopic), zap.Error(err))
			}
			for run == true {
				ev := c.Poll(100)
				switch e := ev.(type) {
				case *kafka.Message:
					users := &golang.AllMatchedUsers{}
					err := proto.Unmarshal(e.Value, users)
					if err != nil {
						//todo update metrics
						logger.Logger.Named(op).Error("Error in unmarshalling message", zap.Error(err))
					}
					//todo create match in database
					//todo publish message: MatchCreated with Id
					//todo send Acknowledgment to publisher
					entityUsers := service.MapToEntityToProtoMessage(users)
					//todo think about context.Background()
					for _, u := range entityUsers {
						if id, err := s.repository.CreateMatch(context.Background(), entity.Game{
							PlayerIDs: u.UserId,
							//Category:  u.Category,
							Status: entity.GameStatusCreated,
						}); err != nil {
							logger.Logger.Named(op).Error("error in creating match", zap.Error(err))
						} else {
							logger.Logger.Named(op).Info(fmt.Sprintf("Publishing Match Created for user: %d", u.UserId), zap.String("matchId", id))
							_, err = s.PublishMatchCreated(request.PublishMatchCreatedRequest{
								UserId:  u.UserId,
								MatchId: id,
							})
							if err != nil {
								logger.Logger.Named(op).Error(
									fmt.Sprintf("Error in publishing MatchId for user: %d", u.UserId),
									zap.Error(err),
									zap.String("matchId", id))
							} else {
								logger.Logger.Named(op).Info(fmt.Sprintf("MatchId published for user: %d", u.UserId), zap.String("matchId", id))
							}
						}
					}
					logger.Logger.Named(op).Info("consumer received message", zap.String("message", fmt.Sprint(entityUsers)), zap.String("time", time.Now().String()))
					// application-specific processing
				case kafka.Error:
					logger.Logger.Named(op).Error("Error in consuming message", zap.Error(e))
					run = false
				default:
					//logger.Logger.Named(op).Error("Unknown message type", zap.Any("message", e))
				}
			}
		}
	default:
		{
			//todo add metrics
			logger.Logger.Named(op).Error("Unhandled type of consumerBroker", zap.Any("consumer", consumer))
		}
	}
	return request.StartMatchCreatorRequest{}, nil
}

func (s Service) PublishMatchCreated(req request.PublishMatchCreatedRequest) (request.PublishMatchCreatedResponse, error) {
	const op = "match.PublishMatchCreated"
	matchCreationTopic := "matchCreated_v1_matchId"
	response := request.PublishMatchCreatedResponse{}
	buff, err := proto.Marshal(&game.GameCreationInfo{
		Id:     req.MatchId,
		UserId: req.UserId,
	})
	if err != nil {
		//todo update metrics
		logger.Logger.Named(op).Error("error in marshaling match message", zap.Error(err))
		return response, err
	}
	producer := s.publisherBroker.Publish(nil)
	switch producer.(type) {
	case *kafka.Producer:
		{
			p := producer.(*kafka.Producer)
			defer p.Close()
			err := p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &matchCreationTopic,
					Partition: kafka.PartitionAny,
				}, Value: buff,
			}, nil)

			if err != nil {
				//todo add metrics
				logger.Logger.Named(op).Error("error in producing message.", zap.String("topic", matchCreationTopic), zap.Error(err))
				return response, err
			}
			return response, nil
		}
	default:
		{
			//todo add metrics
			return response, fmt.Errorf("unhandled type of publisherBroker")
		}
	}
}
