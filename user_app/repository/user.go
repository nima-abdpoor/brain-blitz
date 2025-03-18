package repository

import (
	"BrainBlitz.com/game/adapter/redis"
	entityAuth "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/user_app/service"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

var (
	ErrDuplicateKey = "duplicate key"
)

type Config struct{}

type UserRepository struct {
	Config     Config
	Logger     *slog.Logger
	PostgreSQL *sql.DB
	Cache      *redis.Adapter
}

func NewUserRepository(config Config, db *sql.DB, logger *slog.Logger) UserRepository {
	return UserRepository{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}

func (repo UserRepository) InsertUser(ctx context.Context, user service.User) (int, error) {
	query := "INSERT INTO users (username, password, display_name, role, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id;"

	stmt, err := repo.PostgreSQL.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// should we do this automatically by defining function in postgres or not?
	currentTime := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	var userId int
	err = stmt.QueryRowContext(ctx, user.Username, user.HashedPassword, user.DisplayName, user.Role.String(), currentTime, currentTime).Scan(&userId)
	if err != nil {
		if strings.Contains(err.Error(), ErrDuplicateKey) {
			return 0, fmt.Errorf("duplicate username: %v", err)
		}
		return 0, fmt.Errorf("failed to insert user %v (username: %s)", user.Username, err)
	}

	return userId, nil
}

func (repo UserRepository) GetUser(ctx context.Context, username string) (service.User, error) {
	query := "SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE username = $1 LIMIT 1"

	stmt, err := repo.PostgreSQL.PrepareContext(ctx, query)
	if err != nil {
		return service.User{}, fmt.Errorf("failed to prepare find result by ID statement: %w", err)
	}
	defer stmt.Close()

	var result service.User
	var role string
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	row := stmt.QueryRowContext(ctx, username)
	err = row.Scan(
		&result.ID,
		&result.Username,
		&result.HashedPassword,
		&result.DisplayName,
		&role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, fmt.Errorf("result with username %s not found", username)
		}
		return result, fmt.Errorf("error retrieving user with username: %s, error: %v", username, err)
	}
	result.Role = entityAuth.MapToRoleEntity(role)
	result.CreatedAt = uint64(createdAt.Time.UTC().UnixMilli())
	result.UpdatedAt = uint64(updatedAt.Time.UTC().UnixMilli())
	return result, nil
}

func (repo UserRepository) GetUserById(ctx context.Context, id string) (service.User, error) {
	query := "SELECT id, username, password, display_name, role, created_at, updated_at FROM users WHERE id = $1 LIMIT 1"

	stmt, err := repo.PostgreSQL.PrepareContext(ctx, query)
	if err != nil {
		return service.User{}, fmt.Errorf("failed to prepare find result by ID: %s statement: %w", id, err)
	}
	defer stmt.Close()

	var result service.User
	var role string
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(
		&result.ID,
		&result.Username,
		&result.HashedPassword,
		&result.DisplayName,
		&role,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, fmt.Errorf("result with id %s not found", id)
		}
		return result, fmt.Errorf("error retrieving user with id: %s, error: %v", id, err)
	}
	result.Role = entityAuth.MapToRoleEntity(role)
	result.CreatedAt = uint64(createdAt.Time.UTC().UnixMilli())
	result.UpdatedAt = uint64(updatedAt.Time.UTC().UnixMilli())
	return result, nil
}
