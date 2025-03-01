package user_app

import (
	"BrainBlitz.com/game/adapter/redis"
	"BrainBlitz.com/game/pkg/auth"
	"BrainBlitz.com/game/pkg/grpc"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/postgresql"
	"BrainBlitz.com/game/user_app/repository"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	GRPCServer           grpc.Config       `koanf:"grpc_server"`
	GrpcClient           grpc.Client       `koanf:"grpc_client"`
	PostgresDB           postgresql.Config `koanf:"postgres_db"`
	Repository           repository.Config `koanf:"repository"`
	Redis                redis.Config      `koanf:"redis"`
	Logger               logger.Config     `koanf:"logger"`
	AuthConfig           auth.Config       `koanf:"auth"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}
