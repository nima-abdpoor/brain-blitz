package service

import (
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/email"
)

func (us UserService) SignIn(request *request.SingInRequest) *response.Response {
	if !email.IsValid(request.Email) {
		return us.createFailedResponse(error_code.BadRequest, invalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return us.createFailedResponse(error_code.BadRequest, invalidPasswordErrMsg)
	}

	us.userRepo.GetUser(request.Email)
	utils.CheckPasswordHash(request.Password, "")
	return nil
}
