package repository

import (
	entity "BrainBlitz.com/game/entity/user"
	"errors"
)

var (
	DuplicateUser = errors.New("duplicate user")
)

type UserRepository interface {
	InsertUser(user entity.User) error
	GetUser(email string) (entity.User, error)
	GetUserById(id string) (entity.User, error)
}
