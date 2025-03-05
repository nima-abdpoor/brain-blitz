package main

import (
	"BrainBlitz.com/game/match_app"
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
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
	matchLogger := logger.L()

	matchLogger.Info("match_logger service started...")

	app := match_app.Setup(cfg, matchLogger)
	app.Start()
}
