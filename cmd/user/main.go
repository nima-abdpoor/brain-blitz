package main

import (
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/postgresql"
	"BrainBlitz.com/game/pkg/postgresqlmigrator"
	"BrainBlitz.com/game/user_app"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var cfg user_app.Config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	options := cfgloader.Option{
		Prefix:       "USER_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: filepath.Join(workingDir, "deploy", "user", "development", "config.yaml"),
		CallbackEnv:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load userapp config: %v", err)
	}

	logger.Init(cfg.Logger)
	userLogger := logger.L()

	userLogger.Info("user_app service started...")

	//todo retry to connect in result of connection failure
	//todo add metrics (each connection)
	postgresConn, cnErr := postgresql.Connect(cfg.PostgresDB)
	///////
	// redisConn := redis.Connect(cfg.Redis)
	// cacheManager := redis.NewRedisCache(redisConn, userLogger, cfg.CacheManager)

	if cnErr != nil {
		log.Fatal(cnErr)
	} else {
		userLogger.Info(fmt.Sprintf("You are connected to %s successfully.", cfg.PostgresDB.DBName))
	}

	if err != nil {
		log.Fatalf("Error in Connecting to User Postgresql: %v", err)
	}

	mgr := postgresqlmigrator.New(cfg.PostgresDB, cfg.PostgresDB.PathOfMigration)
	mgr.Up()

	defer postgresql.Close(postgresConn.DB)

	app := user_app.Setup(cfg, postgresConn, userLogger)
	app.Start()
}
