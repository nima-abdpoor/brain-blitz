package userservice

import (
	entity "Game/entity/user"
	"Game/pkg/email"
	"fmt"
)

type Repository interface {
	IsEmailUnique(email string) (bool, error)
	Register(user entity.User) (entity.User, error)
}

type Service struct {
	repo Repository
}

type RegisterRequest struct {
	Email string
	Name  string
}

type RegisterResponse struct {
	User entity.User
}

func (service Service) register(req RegisterRequest) (RegisterResponse, error) {
	// todo verify email

	if !email.IsValid(req.Email) {
		return RegisterResponse{}, fmt.Errorf("email is not Valid")
	}

	if isUnique, err := service.repo.IsEmailUnique(req.Email); err != nil || !isUnique {
		if err != nil {
			return RegisterResponse{}, fmt.Errorf("unexpected error: %w", err)
		}

		if !isUnique {
			return RegisterResponse{}, fmt.Errorf("email is not unique")
		}
	}

	if len(req.Name) <= 3 {
		return RegisterResponse{}, fmt.Errorf("name length should be grater than 3")
	}

	createdUser, err := service.repo.Register(entity.User{
		ID:     0,
		Email:  req.Email,
		Name:   req.Name,
		Avatar: "",
	})

	if err != nil {
		return RegisterResponse{}, fmt.Errorf("unexpected error: %w", err)
	}
	return RegisterResponse{User: createdUser}, nil
}
