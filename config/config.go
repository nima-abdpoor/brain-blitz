package config

import (
	"BrainBlitz.com/game/internal/core/service"
	"BrainBlitz.com/game/internal/infra/repository"
)

type HTTPServer struct {
	Port int `koanf:"port"`
}

type Config struct {
	HTTPServer HTTPServer        `koanf:"http_server"`
	Auth       service.Config    `koanf:"auth"`
	Mysql      repository.Config `koanf:"mysql"`
}
