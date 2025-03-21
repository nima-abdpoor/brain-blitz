package main

import (
	"BrainBlitz.com/game/game_app"
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/mongo"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var cfg game_app.Config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	options := cfgloader.Option{
		Prefix:       "GAME_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: filepath.Join(workingDir, "deploy", "game", "development", "config.yaml"),
		CallbackEnv:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load gameApp config: %v", err)
	}

	logger.Init(cfg.Logger)
	gameLogger := logger.L()

	gameLogger.Info("game service started...")

	connectCtx, cancel := context.WithTimeout(context.Background(), cfg.MongoDB.ConnectTimeout)
	defer cancel()
	mongoDB, err := mongo.NewDB(cfg.MongoDB, connectCtx)
	if err != nil {
		gameLogger.Error(fmt.Sprintf("error in connecting to MongoDB on %s:%d", cfg.MongoDB.Hosts, cfg.MongoDB.Ports), "error", err)
	}

	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.MongoDB.DisconnectTimeout)
		defer cancel()
		err = mongo.Close(mongoDB.DB, disconnectCtx)
		if err != nil {
			gameLogger.Error(fmt.Sprintf("error in disconnecting from MongoDB on %s:%d", cfg.MongoDB.Hosts, cfg.MongoDB.Ports), "error", err)
		}
	}()

	app := game_app.Setup(cfg, mongoDB, gameLogger)
	app.Start()
}
