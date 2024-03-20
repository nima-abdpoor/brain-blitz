package service

import (
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/pkg/email"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"strconv"
)

func (us UserService) SignIn(request *request.SignInRequest) (response.SignInResponse, error) {
	const op = "service.SignIn"
	if !email.IsValid(request.Email) {
		return response.SignInResponse{}, richerror.New(op).
			WithMeta(map[string]interface{}{"email": request.Email}).
			WithMessage(errmsg.InvalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return response.SignInResponse{}, richerror.New(op).
			WithMessage(errmsg.InvalidPasswordErrMsg).
			WithMeta(map[string]interface{}{"password": request.Password})
	}

	if user, err := us.userRepo.GetUser(request.Email); err != nil {
		fmt.Errorf("error In Getting User: %v", err)
		return response.SignInResponse{}, richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		result := utils.CheckPasswordHash(request.Password, user.HashedPassword)
		if result {
			data := make(map[string]string)
			data["user"] = strconv.FormatInt(user.ID, 10)
			accessToken, err := us.authService.CreateAccessToken(data)
			if err != nil {
				return response.SignInResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			refreshToken, err := us.authService.CreateRefreshToken(data)
			if err != nil {
				return response.SignInResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			return response.SignInResponse{
				Username:     user.Username,
				DisplayName:  user.DisplayName,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}, nil
		} else {
			return response.SignInResponse{}, richerror.New(op).
				WithKind(richerror.KindForbidden).
				WithMessage(errmsg.InvalidPasswordErrMsg).
				WithMeta(map[string]interface{}{"password": request})
		}
	}
}
