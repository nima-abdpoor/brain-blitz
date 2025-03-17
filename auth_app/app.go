package auth_app

import (
	"BrainBlitz.com/game/auth_app/delivery/grpc"
	"BrainBlitz.com/game/auth_app/delivery/http"
	"BrainBlitz.com/game/auth_app/service"
	rpc "BrainBlitz.com/game/pkg/grpc"
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
	Service         service.Service
	AuthHandler     http.Handler
	AuthGRPCHandler grpc.Handler
	HTTPServer      http.Server
	GRPCServer      grpc.Server
	Config          Config
	Logger          *slog.Logger
}

func Setup(config Config, logger *slog.Logger) Application {
	authService := service.NewService(config.Service, logger)
	handler := http.NewHandler(authService, logger)
	grpcHandler := grpc.NewHandler(authService, logger)

	return Application{
		Service:     authService,
		AuthHandler: handler,
		HTTPServer:  http.New(httpserver.New(config.HTTPServer), handler, logger),
		GRPCServer:  grpc.NewServer(rpc.New(config.GRPCServer), grpcHandler, logger),
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
	app.Logger.Info("auth_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("HTTP server started on %d", app.Config.HTTPServer.Port))
		if err := app.HTTPServer.Serve(); err != nil {
			// todo add metrics
			app.Logger.Error(fmt.Sprintf("error in HTTP server on %d", app.Config.HTTPServer.Port), "error", err)
		}
		app.Logger.Info(fmt.Sprintf("HTTP server stopped %d", app.Config.HTTPServer.Port))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("GRPC server started on %d", app.Config.GRPCServer.Port))
		if err := app.GRPCServer.Serve(); err != nil {
			//todo add metrics
			app.Logger.Error("error in serving GRPC server user_app listen", err)
		}
		app.Logger.Info(fmt.Sprintf("GRPC server stopped %d", app.Config.GRPCServer.Port))
	}()
}

func (app Application) shutdownServers(ctx context.Context) bool {
	shutdownDone := make(chan struct{})

	go func() {
		var shutdownWg sync.WaitGroup
		shutdownWg.Add(1)
		go app.shutdownHTTPServer(&shutdownWg)

		shutdownWg.Add(1)
		go app.shutdownGRPCServer(&shutdownWg)

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

func (app Application) shutdownGRPCServer(wg *sync.WaitGroup) {
	defer wg.Done()
	app.GRPCServer.Stop()
}
