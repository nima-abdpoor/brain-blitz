package user_app

import (
	"BrainBlitz.com/game/adapter/auth"
	"BrainBlitz.com/game/adapter/redis"
	cachemanager "BrainBlitz.com/game/pkg/cache_manager"
	rpcPkg "BrainBlitz.com/game/pkg/grpc"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/postgresql"
	g "BrainBlitz.com/game/user_app/delivery/grpc"
	"BrainBlitz.com/game/user_app/delivery/http"
	"BrainBlitz.com/game/user_app/repository"
	"BrainBlitz.com/game/user_app/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	Repository   service.Repository
	Service      service.Service
	UserHandler  http.Handler
	HTTPServer   http.Server
	GRPCServer   g.Server
	Config       Config
	Logger       logger.SlogAdapter
	Redis        *redis.Adapter
	CacheManager *cachemanager.CacheManager
}

func Setup(config Config, postgresConn *postgresql.Database, Conn *grpc.ClientConn, logger logger.SlogAdapter) Application {
	redisAdapter := redis.New(config.Redis)
	cache := cachemanager.NewCacheManager(redisAdapter)

	userRepository := repository.NewUserRepository(config.Repository, postgresConn.DB, logger)
	adapter := auth_adapter.New(Conn)
	userService := service.NewService(userRepository, *cache, adapter, logger)
	userHandler := http.NewHandler(userService, logger)
	grpcHandler := g.NewHandler(userService, logger)

	return Application{
		Repository:   userRepository,
		Service:      userService,
		UserHandler:  userHandler,
		HTTPServer:   http.New(httpserver.New(config.HTTPServer), userHandler, logger),
		GRPCServer:   g.New(rpcPkg.New(config.GRPCServer), grpcHandler, logger),
		Config:       config,
		Logger:       logger,
		CacheManager: cache,
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
	app.Logger.Info("user_app stopped")
}

func startServers(app Application, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Logger.Info(fmt.Sprintf("GRPC server started on %d", app.Config.GRPCServer.Port))
		if err := app.GRPCServer.Serve(); err != nil {
			//todo add metrics
			app.Logger.Error("error in serving GRPC server user_app listen", "error", err)
		}
		app.Logger.Info(fmt.Sprintf("GRPC server stopped %d", app.Config.GRPCServer.Port))
	}()

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
