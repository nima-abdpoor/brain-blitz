package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
)

type UserService interface {
	SignUp(request *request.SignUpRequest) (response.SignUpResponse, error)
	SignIn(request *request.SignInRequest) (response.SignInResponse, error)
	Profile(id string) (response.ProfileResponse, error)
}
