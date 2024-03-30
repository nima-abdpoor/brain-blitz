package repository

import (
	entityAuth "BrainBlitz.com/game/entity/auth"
	entity "BrainBlitz.com/game/entity/user"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

type userRepository struct {
	DB repository.Database
}

func NewUserRepository(db repository.Database) repository.UserRepository {
	return &userRepository{
		DB: db,
	}
}

func (ur userRepository) InsertUser(user entity.User) error {
	currentTime := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	if err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		_, err := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Username:    user.Username,
			Password:    user.HashedPassword,
			DisplayName: user.DisplayName,
			Role:        user.Role.String(),
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
		})
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ur userRepository) GetUser(username string) (entity.User, error) {
	const op = "repository.GetUser"
	var result entity.User
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUser(context.Background(), username); err != nil {
			return err
		} else {
			result = entity.User{
				ID:             user.ID,
				Username:       user.Username,
				HashedPassword: user.Password,
				DisplayName:    user.DisplayName,
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

func (ur userRepository) GetUserById(id int64) (entity.User, error) {
	var result entity.User
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUserById(context.Background(), id); err != nil {
			return err
		} else {
			result = entity.User{
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
