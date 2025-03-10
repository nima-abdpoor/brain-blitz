package event

type Topic string

type Event struct {
	Topic   Topic
	Payload []byte
}

type Handler func(event interface{}) error
