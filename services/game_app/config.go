package game_app

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/adapter/redis"
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"BrainBlitz.com/game/adapter/websocket"
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"BrainBlitz.com/game/services/game_app/repository"
	"BrainBlitz.com/game/services/game_app/service"
	"time"
)

type Config struct {
	HTTPServer           httpserver.Config         `koanf:"http_server"`
	Service              service.Config            `koanf:"service"`
	Broker               broker.Config             `koanf:"broker"`
	WebSocket            websocket.Config          `koanf:"websocket"`
	Repository           repository.Config         `koanf:"repository"`
	TaskPublisher        taskqueue.PublisherConfig `koanf:"task_publisher"`
	TaskWorker           taskqueue.WorkerConfig    `koanf:"task_worker"`
	MongoDB              mongo.Config              `koanf:"mongo"`
	Redis                redis.Config              `koanf:"redis"`
	Logger               logger.Config             `koanf:"logger"`
	TotalShutdownTimeout time.Duration             `koanf:"total_shutdown_timeout"`
}
