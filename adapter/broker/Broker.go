package broker

import "context"

type PublisherBroker interface {
	Publish(config map[string]string) any
}

type ConsumerBroker interface {
	Consume(config map[string]string) (any, error)
}

type Config struct {
	Host string
	Port string
}

type Broker interface {
	Publish(ctx context.Context, topic string, message []byte) error
	Consume(ctx context.Context, topic string, handler func([]byte, context.Context) error) error
}
