package main

import (
	"BrainBlitz.com/game/auth_app"
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var cfg auth_app.Config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	options := cfgloader.Option{
		Prefix:       "AUTH_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: filepath.Join(workingDir, "deploy", "auth", "development", "config.yaml"),
		CallbackEnv:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load authApp config: %v", err)
	}

	logger.Init(cfg.Logger)
	authLogger := logger.New()

	authLogger.Info("auth_app service started...")

	app := auth_app.Setup(cfg, authLogger)
	app.Start()
}
