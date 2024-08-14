package match

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/golang/match"
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/logger"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"
	"time"
)

type Service struct {
	repository     repository.MatchManagementRepository
	consumerBroker broker.ConsumerBroker
}

func New(repo repository.MatchManagementRepository, consumer broker.ConsumerBroker) Service {
	return Service{
		repository:     repo,
		consumerBroker: consumer,
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

			}
			for run == true {
				ev := c.Poll(100)
				switch e := ev.(type) {
				case *kafka.Message:
					users := &match.AllMatchedUsers{}
					err := proto.Unmarshal(e.Value, users)
					if err != nil {
						//todo update metrics
						logger.Logger.Named(op).Error("Error in unmarshalling message", zap.Error(err))
					}
					//todo create match in database
					//todo publish message: MatchCreated with Id
					//todo send Acknowledgment to publisher
					entityUsers := match.MapToEntityToProtoMessage(users)
					//todo think about context.Background()
					for _, u := range entityUsers {
						if id, err := s.repository.CreateMatch(context.Background(), entity.Game{
							PlayerIDs: u.UserId,
							Category:  u.Category,
							Status:    entity.GameStatusCreated,
						}); err != nil {
							logger.Logger.Named(op).Error("error in creating match", zap.Error(err))
						} else {
							//todo publish id
							fmt.Println(op, id)
						}
					}
					logger.Logger.Named(op).Info("consumer received message", zap.String("message", fmt.Sprint(entityUsers)), zap.String("time", time.Now().String()))
					// application-specific processing
				case kafka.Error:
					logger.Logger.Named(op).Error("Error in consuming message", zap.Error(e))
					run = false
				default:
					logger.Logger.Named(op).Error("Unknown message type", zap.Any("message", e))
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
