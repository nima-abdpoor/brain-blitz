package service

import (
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/email"
	"fmt"
)

func (us UserService) SignIn(request *request.SingInRequest) *response.Response {
	if !email.IsValid(request.Email) {
		return us.createFailedResponse(error_code.BadRequest, invalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return us.createFailedResponse(error_code.BadRequest, invalidPasswordErrMsg)
	}

	if user, err := us.userRepo.GetUser(request.Email); err != nil {
		fmt.Errorf("error In Getting User: %v", err)
		return &response.Response{
			Data:         nil,
			Status:       false,
			ErrorCode:    error_code.InternalError,
			ErrorMessage: error_code.InternalErrMsg,
		}
	} else {
		result := utils.CheckPasswordHash(request.Password, user.HashedPassword)
		if result {
			data := struct {
				Username    string `json:"username"`
				DisplayName string `json:"displayName"`
				CreatedAt   uint64 `json:"createdAt"`
				UpdatedAt   uint64 `json:"updatedAt"`
			}{
				Username:    user.Username,
				DisplayName: user.DisplayName,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			}
			return &response.Response{
				Data:         data,
				Status:       true,
				ErrorCode:    error_code.Success,
				ErrorMessage: error_code.SuccessErrMsg,
			}
		} else {
			return &response.Response{
				Data:         nil,
				Status:       false,
				ErrorCode:    error_code.BadRequest,
				ErrorMessage: error_code.InvalidPasswordMsg,
			}
		}
	}
}
