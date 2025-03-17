package service

import (
	"BrainBlitz.com/game/metrics"
	"context"
	"github.com/go-co-op/gocron"
	"log/slog"
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
	logger    *slog.Logger
}

func NewScheduler(matchSvc Service, conf SchedulerConfig, logger *slog.Logger) Scheduler {
	return Scheduler{
		scheduler: gocron.NewScheduler(time.UTC),
		service:   matchSvc,
		config:    conf,
		logger:    logger,
	}
}

func (s Scheduler) Start(done <-chan bool) {
	const op = "scheduler.Start"
	s.logger.WithGroup(op).Info("Starting scheduler...")

	if _, err := s.scheduler.Every(s.config.Interval).Second().Do(s.MatchWaitedUsers); err != nil {
		s.logger.WithGroup(op).Error("error in calling MatchWaitedUsers", "error", err.Error())
	}
	s.scheduler.StartAsync()

	<-done
	s.logger.WithGroup(op).Info("stopping scheduler...")
	s.scheduler.Stop()
}

func (s Scheduler) MatchWaitedUsers() {
	const op = "scheduler.MatchWaitedUsers"
	ctx, cancel := context.WithTimeout(context.Background(), s.config.MatchUserTimeOut)
	defer cancel()
	if _, err := s.service.MatchWaitUsers(ctx, MatchWaitedUsersRequest{}); err != nil {
		metrics.FailedMatchedUserCounter.Inc()
		s.logger.WithGroup(op).Error("error in MatchWaitedUsers", "error", err.Error())
	} else {
		metrics.SucceedMatchedUserCounter.Inc()
	}
}
