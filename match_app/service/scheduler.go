package service

import (
	"BrainBlitz.com/game/metrics"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
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
}

func NewScheduler(matchSvc Service, conf SchedulerConfig) Scheduler {
	return Scheduler{
		scheduler: gocron.NewScheduler(time.UTC),
		service:   matchSvc,
		config:    conf,
	}
}

func (s Scheduler) Start(done <-chan bool) {
	const op = "scheduler.Start"
	logger.Logger.Named(op).Info("starting scheduler...")

	if _, err := s.scheduler.Every(s.config.Interval).Second().Do(s.MatchWaitedUsers); err != nil {
		logger.Logger.Named(op).Error("error in calling MatchWaitedUsers", zap.Error(err))
	}
	s.scheduler.StartAsync()

	<-done
	//wait to finish job
	logger.Logger.Named(op).Info("stopping scheduler...")
	s.scheduler.Stop()
}

func (s Scheduler) MatchWaitedUsers() {
	const op = "scheduler.MatchWaitedUsers"
	ctx, cancel := context.WithTimeout(context.Background(), s.config.MatchUserTimeOut)
	defer cancel()
	if _, err := s.service.MatchWaitUsers(ctx, MatchWaitedUsersRequest{}); err != nil {
		metrics.FailedMatchedUserCounter.Inc()
		logger.Logger.Named(op).Error("error in MatchWaitedUsers", zap.Error(err))
	} else {
		metrics.SucceedMatchedUserCounter.Inc()
	}
}
