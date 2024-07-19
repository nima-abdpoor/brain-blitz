package kafka

import (
	"BrainBlitz.com/game/adapter/broker"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
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
		//todo add logger
		log.Printf("error ocured in creating Kafka producer %v\n", err)
	}
	return p
}

func (b Broker) Consume(config map[string]string) (any, error) {
	configMap := kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", b.config.Host, b.config.Port),
	}
	for key, value := range config {
		configMap[key] = value
	}
	consumer, err := kafka.NewConsumer(&configMap)
	if err != nil {
		//todo add metrics
		//todo add logger
		log.Printf("Error ocured in creating comsumer, error:%v", err)
		return nil, err
	}
	return consumer, nil
}
