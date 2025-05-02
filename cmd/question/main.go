package main

import (
	cfgloader "BrainBlitz.com/game/pkg/cfg_loader"
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/pkg/postgresql"
	"BrainBlitz.com/game/services/question_app"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var cfg question_app.Config
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	options := cfgloader.Option{
		Prefix:       "QUESTION_",
		Delimiter:    ".",
		Separator:    "__",
		YamlFilePath: filepath.Join(workingDir, "infra", "deploy", "question", "development", "config.yaml"),
		CallbackEnv:  nil,
	}

	if err := cfgloader.Load(options, &cfg); err != nil {
		log.Fatalf("Failed to load questionapp service: %v", err)
	}

	logger.Init(cfg.Logger)
	questionLogger := logger.New()

	questionLogger.Info("question_app service started...")

	//todo retry to connect in result of connection failure
	//todo add metrics (each connection)
	postgresConn, cnErr := postgresql.Connect(cfg.PostgresDB)

	if cnErr != nil {
		log.Fatal(cnErr)
	} else {
		questionLogger.Info(fmt.Sprintf("You are connected to %s successfully.", cfg.PostgresDB.DBName))
	}

	if err != nil {
		log.Fatalf("Error in Connecting to User Postgresql: %v", err)
	}

	//mgr := postgresqlmigrator.New(cfg.PostgresDB, cfg.PostgresDB.PathOfMigration)
	//mgr.Up()

	defer postgresql.Close(postgresConn.DB)

	app := question_app.Setup(cfg, postgresConn, questionLogger)
	app.Start()
}
