package service

import (
	"BrainBlitz.com/game/metrics"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"github.com/go-co-op/gocron"
	"time"
)

type SchedulerConfig struct {
	Interval         int           `koanf:"interval"`
	MatchUserTimeOut time.Duration `koanf:"match_user_time_out"`
}

type Scheduler struct {
	scheduler *gocron.Scheduler
	service   Service
	config    SchedulerConfig
	logger    logger.SlogAdapter
}

func NewScheduler(matchSvc Service, conf SchedulerConfig, logger logger.SlogAdapter) Scheduler {
	return Scheduler{
		scheduler: gocron.NewScheduler(time.UTC),
		service:   matchSvc,
		config:    conf,
		logger:    logger,
	}
}

func (s Scheduler) Start(done <-chan bool) {
	const op = "scheduler.Start"
	s.logger.Info(op, "message", "starting scheduler")

	if _, err := s.scheduler.Every(s.config.Interval).Second().Do(s.MatchWaitedUsers); err != nil {
		s.logger.Error(op, "message", "error in calling MatchWaitedUsers", "error", err.Error())
	}
	s.scheduler.StartAsync()

	<-done
	s.logger.Info(op, "message", "stopping scheduler")
	s.scheduler.Stop()
}

func (s Scheduler) MatchWaitedUsers() {
	const op = "scheduler.MatchWaitedUsers"
	ctx, cancel := context.WithTimeout(context.Background(), s.config.MatchUserTimeOut)
	defer cancel()
	if _, err := s.service.MatchWaitUsers(ctx, MatchWaitedUsersRequest{}); err != nil {
		metrics.FailedMatchedUserCounter.Inc()
		s.logger.Error(op, "message", "error in MatchWaitedUsers", "error", err.Error())
	} else {
		metrics.SucceedMatchedUserCounter.Inc()
	}
}
