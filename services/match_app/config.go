package match_app

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/pkg/grpc"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/services/match_app/repository"
	"BrainBlitz.com/game/services/match_app/service"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config       `koanf:"http_server"`
	GRPCServer           grpc.Config             `koanf:"grpc_server"`
	Broker               broker.Config           `koanf:"broker"`
	Repository           repository.Config       `koanf:"repository"`
	Service              service.Config          `koanf:"service"`
	Scheduler            service.SchedulerConfig `koanf:"scheduler"`
	Redis                redis.Config            `koanf:"redis"`
	Logger               logger.Config           `koanf:"logger"`
	TotalShutdownTimeout time.Duration           `koanf:"total_shutdown_timeout"`
}
