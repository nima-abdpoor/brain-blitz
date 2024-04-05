package repository

import (
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/config"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

func (db database) GetDB() *sqlc.Queries {
	return sqlc.New(db)
}

func (db database) Close() error {
	return db.DB.Close()
}

// ExecTx Execute a function within a database transaction.
func (db *database) ExecTx(ctx context.Context, fn func(queries *sqlc.Queries) error) error {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type Config struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Port     int    `koanf:"port"`
	Host     string `koanf:"host"`
	DBName   string `koanf:"db_name"`
}
