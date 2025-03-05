package match_app

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/match_app/repository"
	"BrainBlitz.com/game/match_app/service"
	"BrainBlitz.com/game/pkg/grpc"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	GRPCServer           grpc.Config       `koanf:"grpc_server"`
	Repository           repository.Config `koanf:"repository"`
	Service              service.Config    `koanf:"service"`
	Redis                redis.Config      `koanf:"redis"`
	Logger               logger.Config     `koanf:"logger"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}
