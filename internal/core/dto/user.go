package dto

type UserDTO struct {
	Username       string
	HashedPassword string
	DisplayName    string
	CreatedAt      uint64
	UpdatedAt      uint64
}
