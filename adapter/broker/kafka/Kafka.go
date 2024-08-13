package kafka

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/logger"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type Config struct {
	Host string `koanf:"host"`
	Port string `koanf:"port"`
}

type Broker struct {
	config Config
}

func NewKafkaPublisher(config Config) broker.PublisherBroker {
	return Broker{
		config: config,
	}
}

func NewKafkaConsumer(config Config) broker.ConsumerBroker {
	return Broker{
		config: config,
	}
}

func (b Broker) Publish(config map[string]string) any {
	const op = "kafka.Publish"

	//todo refactor to using multiple address instead of single node
	configMap := kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", b.config.Host, b.config.Port),
	}
	for key, value := range config {
		configMap[key] = value
	}
	p, err := kafka.NewProducer(&configMap)
	if err != nil {
		//todo add metrics
		logger.Logger.Named(op).Error("error occurred in creating Kafka producer", zap.Error(err))
	}
	return p
}

func (b Broker) Consume(config map[string]string) (any, error) {
	const op = "kafka.Consume"
	configMap := kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", b.config.Host, b.config.Port),
	}
	for key, value := range config {
		configMap[key] = value
	}
	consumer, err := kafka.NewConsumer(&configMap)
	if err != nil {
		//todo add metrics
		logger.Logger.Named(op).Error("Error occurred in creating consumer", zap.Error(err))
		return nil, err
	}
	return consumer, nil
}
