package broker

type PublisherBroker interface {
	Publish(config map[string]string) any
}

type ConsumerBroker interface {
	Consume(config map[string]string) (any, error)
}
