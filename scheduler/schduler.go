package scheduler

import (
	"fmt"
	"time"
)

type Scheduler struct {
}

func New() Scheduler {
	return Scheduler{}
}

func (Scheduler Scheduler) Start(done <-chan bool) {
	for {
		select {
		case d := <-done:
			fmt.Println("scheduler exiting...", d)
			return
		default:
			now := time.Now()
			fmt.Println("scheduler now", now)
			time.Sleep(5 * time.Second)
		}
	}
}
