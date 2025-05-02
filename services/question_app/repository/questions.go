package repository

import (
	"BrainBlitz.com/game/pkg/logger"
	"database/sql"
)

type Config struct{}

type QuestionRepository struct {
	Config     Config
	Logger     logger.Logger
	PostgreSQL *sql.DB
}

func New(config Config, db *sql.DB, logger logger.Logger) QuestionRepository {
	return QuestionRepository{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}
