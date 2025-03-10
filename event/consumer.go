package event

import (
	"fmt"
	"log/slog"
)

type handlerFunc func(event Event) error
type Router map[Topic]handlerFunc

type EventConsumer struct {
	Consumers []Consumer
	Router    Router
	Logger    *slog.Logger
}

func (c EventConsumer) Start(done <-chan bool) {
	eventStream := make(chan Event, 1024)
	for _, consumer := range c.Consumers {
		err := consumer.Consume(eventStream)
		if err != nil {
			c.Logger.Error("can't start consuming events", "error", err)
		}
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case e := <-eventStream:
				err := c.Router[e.Topic](e)
				if err != nil {
					c.Logger.Error(fmt.Sprintf("can't handle event with topic: %s", e.Topic), "error", err)
				}
			}
		}
	}()
}
