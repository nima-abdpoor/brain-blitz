package service

// todo move this to somewhere general
import entity "BrainBlitz.com/game/entity/auth"

type User struct {
	ID             int64
	Username       string
	HashedPassword string
	DisplayName    string
	CreatedAt      uint64
	UpdatedAt      uint64
	Role           entity.Role
}
