package repository

import (
	"BrainBlitz.com/game/internal/core/dto"
	"errors"
)

var (
	DuplicateUser = errors.New("duplicate user")
)

type UserRepository interface {
	InsertUser(dto dto.UserDTO) error
	GetUser(email string) (dto.UserDTO, error)
}
