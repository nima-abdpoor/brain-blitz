package match_app

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/match_app/delivery/http"
	"BrainBlitz.com/game/match_app/repository"
	"BrainBlitz.com/game/match_app/service"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"log/slog"
)

type Application struct {
	Repository  service.Repository
	Service     service.Service
	UserHandler http.Handler
	HTTPServer  http.Server
}

func Setup(config Config, logger *slog.Logger) Application {
	redisAdapter := redis.New(config.Redis)
	repo := repository.NewRepository(config.Repository, logger, redisAdapter)
	srv := service.NewService(repo, config.Service)
	handler := http.NewHandler(srv)

	return Application{
		Repository:  repo,
		Service:     srv,
		UserHandler: handler,
		HTTPServer:  http.New(httpserver.New(config.HTTPServer), handler, logger),
	}
}
