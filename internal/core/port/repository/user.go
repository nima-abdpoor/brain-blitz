package repository

import (
	"BrainBlitz.com/game/internal/core/dto"
	"errors"
)

var (
	DuplicateUser = errors.New("duplicate user")
)

type UserRepository interface {
	Insert(dto dto.UserDTO) error
}
