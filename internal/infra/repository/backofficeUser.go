package repository

import (
	entityAuth "BrainBlitz.com/game/entity/auth"
	entity "BrainBlitz.com/game/entity/user"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/infra/repository/sqlc"
	"context"
	"log"
)

type BackofficeUser struct {
	DB repository.Database
}

func New(db repository.Database) repository.BackofficeUserRepository {
	return &BackofficeUser{
		DB: db,
	}
}

func (backofficeUser BackofficeUser) ListUsers() ([]entity.User, error) {
	// todo declaring empty slice to avoid null in result.
	users := []entity.User{}
	err := backofficeUser.DB.ExecTx(context.Background(), func(queries *sqlc.Queries) error {
		if result, err := queries.GetUsers(context.Background()); err != nil {
			log.Println(err)
			return err
		} else {
			for _, user := range result {
				users = append(users, entity.User{
					ID:          user.ID,
					Username:    user.Username,
					DisplayName: user.DisplayName,
					CreatedAt:   uint64(user.CreatedAt.Time.UTC().UnixMilli()),
					UpdatedAt:   uint64(user.UpdatedAt.Time.UTC().UnixMilli()),
					Role:        entityAuth.MapToRoleEntity(user.Role),
				})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}
