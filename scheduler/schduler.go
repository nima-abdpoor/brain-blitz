package scheduler

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"fmt"
	"github.com/go-co-op/gocron"
	"sync"
	"time"
)

type Scheduler struct {
	sch      *gocron.Scheduler
	matchSvc service.MatchMakingService
	conf     Config
}

type Config struct {
	Interval int `koanf:"interval"`
}

func New(matchSvc service.MatchMakingService, conf Config) Scheduler {
	return Scheduler{
		sch:      gocron.NewScheduler(time.UTC),
		matchSvc: matchSvc,
		conf:     conf,
	}
}

func (s Scheduler) Start(done <-chan bool, wg *sync.WaitGroup) {
	const op = "scheduler.Start"

	defer wg.Done()

	if _, err := s.sch.Every(s.conf.Interval).Second().Do(s.MatchWaitedUsers); err != nil {
		fmt.Println(op, err)
	}
	s.sch.StartAsync()

	<-done
	//wait to finish job
	fmt.Println("stopping scheduler...")
	s.sch.Stop()
}

func (s Scheduler) MatchWaitedUsers() {
	_, _ = s.matchSvc.MatchWaitUsers(&request.MatchWaitedUsersRequest{})
}
