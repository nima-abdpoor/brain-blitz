package scheduler

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/port/service"
	"context"
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
	Interval         int           `koanf:"interval"`
	MatchUserTimeOut time.Duration `koanf:"match_user_time_out"`
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
	fmt.Println(op, "scheduler started...")
	defer wg.Done()

	go s.matchSvc.StartMatchMaker()
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
	const op = "scheduler.MatchWaitedUsers"
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.MatchUserTimeOut)
	defer cancel()
	if _, err := s.matchSvc.MatchWaitUsers(ctx, &request.MatchWaitedUsersRequest{}); err != nil {
		fmt.Println(op, err)
	}
}
