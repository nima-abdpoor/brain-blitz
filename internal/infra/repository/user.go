package repository

import (
	"BrainBlitz.com/game/internal/core/dto"
	"BrainBlitz.com/game/internal/core/port/repository"
)

type UserRepository struct {
	DB repository.Database
}

func (ur UserRepository) Insert(dto dto.UserDTO) error {

}
