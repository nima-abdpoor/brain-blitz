package config

import (
	"BrainBlitz.com/game/feature"
	"BrainBlitz.com/game/internal/core/server/http"
	"BrainBlitz.com/game/internal/core/service"
	presenceService "BrainBlitz.com/game/internal/core/service/presence"
	"BrainBlitz.com/game/internal/infra/repository"
	"BrainBlitz.com/game/internal/infra/repository/presence"
	"BrainBlitz.com/game/internal/infra/repository/redis"
)

type Config struct {
	Feature     feature.Config         `koanf:"feature"`
	HTTPServer  http.Config            `koanf:"http"`
	Auth        service.Config         `koanf:"auth"`
	Mysql       repository.Config      `koanf:"mysql"`
	Redis       redis.Config           `koanf:"redis"`
	Presence    presenceService.Config `koanf:"presence_service"`
	GetPresence presence.Config        `koanf:"presence"`
}
