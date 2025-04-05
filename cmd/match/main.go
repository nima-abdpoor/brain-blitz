package main

import (
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/services/match_app"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var cfg match_app.Config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	options := cfgloader.Option{
		Prefix:       "MATCH_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: filepath.Join(workingDir, "deploy", "match", "development", "config.yaml"),
		CallbackEnv:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load matchApp config: %v", err)
	}

	logger.Init(cfg.Logger)
	matchLogger := logger.New()

	matchLogger.Info("match service started...")

	app := match_app.Setup(cfg, matchLogger)
	app.Start()
}
