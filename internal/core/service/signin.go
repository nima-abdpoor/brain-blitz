package service

import (
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/logger"
	"BrainBlitz.com/game/pkg/email"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"go.uber.org/zap"
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
		logger.Logger.Named(op).Error("error In Getting User", zap.String("email", request.Email), zap.Error(err))
		return response.SignInResponse{}, err
	} else {
		result := utils.CheckPasswordHash(request.Password, user.HashedPassword)
		if result {
			data := make(map[string]string)
			data["user"] = strconv.FormatInt(user.ID, 10)
			data["role"] = user.Role.String()
			accessToken, err := us.authService.CreateAccessToken(data)
			if err != nil {
				// todo add metrics
				logger.Logger.Named(op).Error("error creating Access Token", zap.Error(err))
				return response.SignInResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			refreshToken, err := us.authService.CreateRefreshToken(data)
			if err != nil {
				logger.Logger.Named(op).Error("error In Creating Refresh Token", zap.String("data", fmt.Sprint(data)), zap.Error(err))
				return response.SignInResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			return response.SignInResponse{
				ID:           strconv.FormatInt(user.ID, 10),
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
