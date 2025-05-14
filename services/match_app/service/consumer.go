package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"sync"
)

type Consumer struct {
	broker  broker.Broker
	service Service
	logger  logger.SlogAdapter
}

func NewConsumer(broker broker.Broker, service Service, logger logger.SlogAdapter) Consumer {
	return Consumer{
		broker:  broker,
		service: service,
		logger:  logger,
	}
}

func (c Consumer) Consume() {
	wg := &sync.WaitGroup{}
	topics := c.getTopics()
	wg.Add(len(topics))
	for topic, handler := range topics {
		go func() {
			defer wg.Done()
			ctx := context.WithoutCancel(context.Background())
			err := c.broker.Consume(ctx, topic, handler)
			if err != nil {
				c.logger.Error("error in consuming", "topic", topic, "error", err)
			}
		}()
	}
}

func (c Consumer) getTopics() map[string]func([]byte, context.Context) error {
	var topics = []string{
		event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED,
	}

	result := make(map[string]func([]byte, context.Context) error)
	for _, topic := range topics {
		switch topic {
		case event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED:
			result[topic] = c.service.ConsumeJoinMatchQueueRequest
		default:
			c.logger.Warn("Unknown topic", "topic", topic)
			return result
		}

	}
	return result
}
