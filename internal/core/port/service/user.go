package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
)

type UserService interface {
	SignUp(request *request.SignUpRequest) *response.Response
	SignIn(request *request.SignInRequest) (response.SignInResponse, error)
}
