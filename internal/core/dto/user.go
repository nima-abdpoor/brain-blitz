package dto

type UserDTO struct {
	Email          string
	HashedPassword string
	DisplayName    string
	CreatedAt      uint64
	UpdatedAt      uint64
}
