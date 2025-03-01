package repository

import (
	entityAuth "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/infra/repository/redis"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"BrainBlitz.com/game/user_app/service"
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"
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

func (ur UserRepository) InsertUser(user service.User) error {
	currentTime := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	if err := ur.PostgreSQL.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		_, err := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Username:    user.Username,
			Password:    user.HashedPassword,
			DisplayName: user.DisplayName,
			Role:        user.Role.String(),
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
		})
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ur UserRepository) GetUser(username string) (service.User, error) {
	const op = "repository.GetUser"
	var result service.User
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUser(context.Background(), username); err != nil {
			return err
		} else {
			result = service.User{
				ID:             user.ID,
				Username:       user.Username,
				HashedPassword: user.Password,
				DisplayName:    user.DisplayName,
				Role:           entityAuth.MapToRoleEntity(user.Role),
				CreatedAt:      uint64(user.CreatedAt.Time.UTC().UnixMilli()),
				UpdatedAt:      uint64(user.CreatedAt.Time.UTC().UnixMilli()),
			}
		}
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return result, richerror.New(op).WithKind(richerror.KindNotFound).WithMessage(errmsg.UserNotFound)
		}
		return result, richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	}
	return result, nil
}

func (ur UserRepository) GetUserById(id int64) (service.User, error) {
	var result service.User
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUserById(context.Background(), id); err != nil {
			return err
		} else {
			result = service.User{
				ID:             user.ID,
				Username:       user.Username,
				HashedPassword: user.Password,
				DisplayName:    user.DisplayName,
				Role:           entityAuth.MapToRoleEntity(user.Role),
				CreatedAt:      uint64(user.CreatedAt.Time.UTC().UnixMilli()),
				UpdatedAt:      uint64(user.UpdatedAt.Time.UTC().UnixMilli()),
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}
