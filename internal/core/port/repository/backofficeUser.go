package repository

import entity "BrainBlitz.com/game/entity/user"

type BackofficeUserRepository interface {
	ListUsers() ([]entity.User, error)
}
