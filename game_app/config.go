package game_app

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/game_app/repository"
	"BrainBlitz.com/game/game_app/service"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	Service              service.Config    `koanf:"service"`
	WebSocket            websocket.Config  `koanf:"websocket"`
	Repository           repository.Config `koanf:"repository"`
	MongoDB              mongo.Config      `koanf:"mongo"`
	Logger               logger.Config     `koanf:"logger"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}
