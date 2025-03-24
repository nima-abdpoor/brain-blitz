package repository

import (
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"context"
	"io"
)

type Database interface {
	io.Closer
	GetDB() *sqlc.Queries
	ExecTx(ctx context.Context, fn func(queries *sqlc.Queries) error) error
}
