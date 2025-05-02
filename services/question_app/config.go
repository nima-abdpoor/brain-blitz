package question_app

import (
	"BrainBlitz.com/game/adapter/broker"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/postgresql"
	"BrainBlitz.com/game/services/question_app/repository"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config `koanf:"http_server"`
	PostgresDB           postgresql.Config `koanf:"postgres_db"`
	Repository           repository.Config `koanf:"repository"`
	Broker               broker.Config     `koanf:"broker"`
	Logger               logger.Config     `koanf:"logger"`
	TotalShutdownTimeout time.Duration     `koanf:"total_shutdown_timeout"`
}
