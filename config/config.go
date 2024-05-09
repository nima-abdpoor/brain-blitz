package config

import (
	"BrainBlitz.com/game/internal/core/service"
	matchMakingHandler "BrainBlitz.com/game/internal/core/service/matchMaking"
	presenceService "BrainBlitz.com/game/internal/core/service/presence"
	"BrainBlitz.com/game/internal/infra/repository"
	"BrainBlitz.com/game/internal/infra/repository/matchmaking"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/scheduler"
)

type HTTPServer struct {
	Port int `koanf:"port"`
}

type Config struct {
	HTTPServer         HTTPServer                `koanf:"http_server"`
	Auth               service.Config            `koanf:"auth"`
	Mysql              repository.Config         `koanf:"mysql"`
	Redis              redis.Config              `koanf:"redis"`
	MatchMakingPrefix  matchmaking.Config        `koanf:"matchMaking"`
	MatchMakingTimeOut matchMakingHandler.Config `koanf:"matchMaking"`
	Presence           presenceService.Config    `koanf:"presence_service"`
	Scheduler          scheduler.Config          `koanf:"scheduler"`
}
