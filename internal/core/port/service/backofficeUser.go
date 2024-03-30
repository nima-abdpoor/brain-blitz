package service

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
)

type BackofficeUserService interface {
	ListUsers(request *request.ListUserRequest) (response.ListUserResponse, error)
}
