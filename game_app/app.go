package game_app

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/game_app/delivery/http"
	"BrainBlitz.com/game/game_app/repository"
	"BrainBlitz.com/game/game_app/service"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	Repository  service.Repository
	Service     service.Service
	UserHandler http.Handler
	Consumer    service.Consumer
	HTTPServer  http.Server
	Config      Config
	Logger      logger.SlogAdapter
}

func Setup(config Config, db *mongo.Database, logger logger.SlogAdapter) Application {
	gameRepository := repository.NewGameRepository(config.Repository, logger, db)
	ws := websocket.NewWS(config.WebSocket)
	gameService := service.NewService(config.Service, gameRepository, ws, logger)

	kafkaBroker, err := broker.NewKafkaBroker([]string{fmt.Sprintf("%s:%s", config.Broker.Host, config.Broker.Port)}, logger)
	if err != nil {
		logger.Error("Error creating kafka broker", "error", err)
		panic(err)
	}

	consumer := service.NewConsumer(kafkaBroker, gameService, logger)
	userHandler := http.NewHandler(gameService, logger)

	return Application{
		Repository:  gameRepository,
		Service:     gameService,
		UserHandler: userHandler,
		Consumer:    consumer,
		HTTPServer:  http.New(httpserver.New(config.HTTPServer), userHandler, logger),
		Config:      config,
		Logger:      logger,
	}
}

func (app Application) Start() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startServers(app, &wg)
	<-ctx.Done()
	app.Logger.Info("Shutdown signal received...")

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), app.Config.TotalShutdownTimeout)
	defer cancel()

	if app.shutdownServers(shutdownTimeoutCtx) {
		app.Logger.Info("Servers shut down gracefully")
	} else {
		app.Logger.Warn("Shutdown timed out, exiting application")
		os.Exit(1)
	}

	wg.Wait()
	app.Logger.Info("match_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
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
		app.Logger.Info("Consumer Started")
		app.Consumer.Consume()
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(&shutdownWg)

		shutdownWg.Wait()
		close(shutdownDone)
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
