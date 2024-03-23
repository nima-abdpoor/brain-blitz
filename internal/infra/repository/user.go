package repository

import (
	entity "BrainBlitz.com/game/entity/user"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"context"
	"database/sql"
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

func (ur userRepository) GetUser(username string) (entity.User, error) {
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
		return result, err
	}
	return result, nil
}

func (ur userRepository) GetUserById(id string) (entity.User, error) {
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
