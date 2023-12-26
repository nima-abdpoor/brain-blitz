package repository

import "BrainBlitz.com/game/internal/core/dto"

type UserRepository interface {
	Insert(dto *dto.UserDTO) error
}
