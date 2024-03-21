package repository

import (
	"BrainBlitz.com/game/internal/core/dto"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	duplicateEntryMsg = "Duplicate entry"
	numberRowInserted = 1
)

type userRepository struct {
	DB repository.Database
}

func NewUserRepository(db repository.Database) repository.UserRepository {
	return &userRepository{
		DB: db,
	}
}

func (ur userRepository) InsertUser(dto dto.UserDTO) error {
	currentTime := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	if err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		res, err := queries.CreateUser(context.Background(), sqlc.CreateUserParams{
			Username:    dto.Username,
			Password:    dto.HashedPassword,
			DisplayName: dto.DisplayName,
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
		})
		if err != nil {
			return err
		}
		fmt.Println(res)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ur userRepository) GetUser(username string) (dto.UserDTO, error) {
	var result dto.UserDTO
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUser(context.Background(), username); err != nil {
			return err
		} else {
			result = dto.UserDTO{
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

func (ur userRepository) GetUserById(id string) (dto.UserDTO, error) {
	var result dto.UserDTO
	err := ur.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if user, err := queries.GetUserById(context.Background(), id); err != nil {
			return err
		} else {
			result = dto.UserDTO{
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
