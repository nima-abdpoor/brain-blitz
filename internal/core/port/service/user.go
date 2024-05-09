package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"context"
)

type UserService interface {
	SignUp(request *request.SignUpRequest) (response.SignUpResponse, error)
	SignIn(request *request.SignInRequest) (response.SignInResponse, error)
	Profile(ctx context.Context, id int64) (response.ProfileResponse, error)
}
