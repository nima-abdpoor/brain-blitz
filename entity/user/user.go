package entity

type User struct {
	ID             int64
	Username       string
	HashedPassword string
	DisplayName    string
	CreatedAt      uint64
	UpdatedAt      uint64
}
