package auth_app

import (
	"BrainBlitz.com/game/auth_app/service"
	"BrainBlitz.com/game/pkg/grpc"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	GRPCServer           grpc.Config       `koanf:"grpc_server"`
	Service              service.Config    `koanf:"service"`
	Logger               logger.Config     `koanf:"logger"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}
