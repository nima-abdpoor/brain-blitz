package matchMakingHandler

import (
	"BrainBlitz.com/game/contract/golang/match"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

func (s Service) StartMatchMaker() {
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
						log.Printf("%s Error in unmarshaling message: %v\n", op, e)
					}
					entityUsers := match.MapToEntityToProtoMessage(users)
					log.Printf("%s, value of consumer is:%s time:%s\n\n", op, entityUsers, time.Now().String())
					// application-specific processing
				case kafka.Error:
					log.Printf("%s Error in consuming message: %v\n", op, e)
					run = false
				default:
					//log.Printf("%s Ignoring this uknow type %v\n", op, e)
				}
			}
		}
	default:
		{
			//todo add metrics
			//todo add logger
			log.Printf("Unhandled type of consumerBroker %s", consumer)
		}
	}
}
