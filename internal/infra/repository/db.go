package repository

import (
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/config"
	"database/sql"
	"time"
)

type database struct {
	*sql.DB
}

func NewDB(conf config.DatabaseConfig) (repository.Database, error) {
	db, err := newDatabase(conf)
	if err != nil {
		return nil, err
	}
	return &database{db}, nil
}

func newDatabase(conf config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open(conf.Driver, conf.Url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(conf.MaxOpenCons)
	db.SetMaxIdleConns(conf.MaxIdleCons)
	db.SetConnMaxLifetime(time.Minute * time.Duration(conf.ConnMaxLifeTimeMinutes))

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}

func (db database) GetDB() *sql.DB {
	return db.DB
}

func (db database) Close() error {
	return db.DB.Close()
}
