package match_app

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/match_app/delivery/http"
	"BrainBlitz.com/game/match_app/repository"
	"BrainBlitz.com/game/match_app/service"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	Config      Config
	Broker      broker.Broker
	Repository  service.Repository
	Service     service.Service
	Scheduler   service.Scheduler
	UserHandler http.Handler
	HTTPServer  http.Server
	Logger      *slog.Logger
}

func Setup(config Config, logger *slog.Logger) Application {
	redisAdapter := redis.New(config.Redis)
	repo := repository.NewRepository(config.Repository, logger, redisAdapter)
	kafkaBroker, err := broker.NewKafkaBroker([]string{fmt.Sprintf("%s:%s", config.Broker.Host, config.Broker.Port)}, logger)
	svc := service.NewService(repo, config.Service, kafkaBroker, logger)
	scheduler := service.NewScheduler(svc, config.Scheduler, logger)
	handler := http.NewHandler(svc)
	if err != nil {
		logger.Error("Error creating kafka broker", "error", err)
		panic(err)
	}

	return Application{
		Config:      config,
		Repository:  repo,
		Service:     svc,
		UserHandler: handler,
		Scheduler:   scheduler,
		HTTPServer:  http.New(httpserver.New(config.HTTPServer), handler, logger),
		Logger:      logger,
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup
	schedulerDoneChan := make(chan bool)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, schedulerDoneChan, &wg)
	<-ctx.Done()
	app.Logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx, schedulerDoneChan) {
		app.Logger.Info("Servers shut down gracefully")
	} else {
		app.Logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	app.Logger.Info("match_app stopped")
}

func startServers(app Application, done <-chan bool, wg *sync.WaitGroup) {
	wg.Add(2)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			// todo add metrics
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), "error", err)
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()
	go func() {
		defer wg.Done()

		app.Logger.Info("Scheduler Started")
		app.Scheduler.Start(done)
	}()
}

func (app Application) shutdownServers(ctx context.Context, done chan<- bool) bool {
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(&shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
	}()

	go func() {
		done <- true
	}()

	select {
	case <-shutdownDone:
		return true
	case <-ctx.Done():
		return false
	}
}

func (app Application) shutdownHTTPServer(wg *sync.WaitGroup) {
	defer wg.Done()
	httpShutdownCtx, httpCancel := context.WithTimeout(context.Background(), app.Config.HTTPServer.ShutDownCtxTimeout)
	defer httpCancel()
	if err := app.HTTPServer.Stop(httpShutdownCtx); err != nil {
		app.Logger.Error(fmt.Sprintf("HTTP server graceful shutdown failed: %v", err))
	}
}
