package broker

import (
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"fmt"
	"github.com/IBM/sarama"
)

type KafkaBroker struct {
	Logger   logger.SlogAdapter
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func NewKafkaBroker(brokers []string, logger logger.SlogAdapter) (*KafkaBroker, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &KafkaBroker{producer: producer, consumer: consumer, Logger: logger}, nil
}

func (k *KafkaBroker) Publish(ctx context.Context, topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := k.producer.SendMessage(msg)
	if err != nil {
		k.Logger.Error(fmt.Sprintf("Failed to publish message to Kafka topic %s", topic), "error", err)
		return err
	}

	k.Logger.Info("Published to Kafka", "topic", topic, "message", string(message))
	return nil
}

func (k *KafkaBroker) Consume(ctx context.Context, topic string, handler func([]byte, context.Context) error) error {
	partitionConsumer, err := k.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			if err := handler(msg.Value, ctx); err != nil {
				k.Logger.Error("Error handling message from Kafka", "error", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (k *KafkaBroker) Close() error {
	if err := k.producer.Close(); err != nil {
		return err
	}
	if err := k.consumer.Close(); err != nil {
		return err
	}
	return nil
}
