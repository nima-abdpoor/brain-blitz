package service

import (
	authEntity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/logger"
	cachemanager "BrainBlitz.com/game/pkg/cache_manager"
	utils2 "BrainBlitz.com/game/pkg/common"
	"BrainBlitz.com/game/pkg/email"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type Repository interface {
	InsertUser(ctx context.Context, user User) (int, error)
	GetUser(ctx context.Context, email string) (User, error)
	GetUserById(ctx context.Context, id int64) (User, error)
}

type Service struct {
	repository   Repository
	CacheManager cachemanager.CacheManager
}

func NewService(repository Repository, cm cachemanager.CacheManager) Service {
	return Service{
		repository:   repository,
		CacheManager: cm,
	}
}

func (s Service) SignUp(ctx context.Context, request SignUpRequest) (SignUpResponse, error) {
	const op = "service.SignUp"
	if !email.IsValid(request.Email) {
		return SignUpResponse{}, richerror.New(op).
			WithMeta(map[string]interface{}{"email": request.Email}).
			WithMessage(errmsg.InvalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return SignUpResponse{}, richerror.New(op).
			WithMessage(errmsg.InvalidPasswordErrMsg).
			WithMeta(map[string]interface{}{"password": request.Password})
	}

	currentTime := utils2.GetUTCCurrentMillis()

	hashPassword, err := utils2.HashPassword(request.Password)
	if err != nil {
		return SignUpResponse{}, richerror.New(op).
			WithKind(richerror.KindUnexpected).
			WithMeta(map[string]interface{}{"ERROR_CODE": error_code.BcryptErrorHashingPassword})
	}

	userDto := User{
		Username:       request.Email,
		HashedPassword: hashPassword,
		DisplayName:    getDisplayName(request.Email),
		Role:           authEntity.UserRole,
		CreatedAt:      currentTime,
		UpdatedAt:      currentTime,
	}

	_, err = s.repository.InsertUser(ctx, userDto)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return SignUpResponse{}, richerror.New(op).
				WithError(err).
				WithKind(richerror.KindInvalid).
				WithMessage(errmsg.DuplicateUsername)
		}
		//todo add to metrics
		logger.Logger.Named(op).Error("Error in inserting User", zap.String("userDto", fmt.Sprint(userDto)), zap.Error(err))
		return SignUpResponse{}, richerror.New(op).
			WithError(err).
			WithKind(richerror.KindUnexpected)
	}

	return SignUpResponse{
		DisplayName: userDto.DisplayName,
	}, nil
}

func (s Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {
	const op = "service.Login"
	if !email.IsValid(request.Email) {
		return LoginResponse{}, richerror.New(op).
			WithMeta(map[string]interface{}{"email": request.Email}).
			WithMessage(errmsg.InvalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return LoginResponse{}, richerror.New(op).
			WithMessage(errmsg.InvalidPasswordErrMsg).
			WithMeta(map[string]interface{}{"password": request.Password})
	}

	if user, err := s.repository.GetUser(ctx, request.Email); err != nil {
		logger.Logger.Named(op).Error("error In Getting User", zap.String("email", request.Email), zap.Error(err))
		return LoginResponse{}, err
	} else {
		result := utils2.CheckPasswordHash(request.Password, user.HashedPassword)
		if result {
			data := make(map[string]string)
			data["user"] = strconv.FormatInt(user.ID, 10)
			data["role"] = user.Role.String()
			//todo it should call rpc to authService
			//accessToken, err := s.authService.CreateAccessToken(data)
			accessToken := "aaa"
			if err != nil {
				// todo add metrics
				logger.Logger.Named(op).Error("error creating Access Token", zap.Error(err))
				return LoginResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			//todo it should call rpc to authService
			//refreshToken, err := s.authService.CreateRefreshToken(data)
			refreshToken := "rrr"
			if err != nil {
				logger.Logger.Named(op).Error("error In Creating Refresh Token", zap.String("data", fmt.Sprint(data)), zap.Error(err))
				return LoginResponse{}, richerror.New(op).
					WithKind(richerror.KindUnexpected).
					WithError(err).
					WithMeta(map[string]interface{}{"data": data})
			}
			return LoginResponse{
				ID:           strconv.FormatInt(user.ID, 10),
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			}, nil
		} else {
			return LoginResponse{}, richerror.New(op).
				WithKind(richerror.KindForbidden).
				WithMessage(errmsg.InvalidPasswordErrMsg).
				WithMeta(map[string]interface{}{"password": request})
		}
	}
}

func (s Service) Profile(ctx context.Context, request ProfileRequest) (ProfileResponse, error) {
	const op = "service.Profile"
	if user, err := s.repository.GetUserById(ctx, request.ID); err != nil {
		// todo check if logger needed
		// todo add metrics
		return ProfileResponse{}, richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		return ProfileResponse{
			ID:          strconv.FormatInt(user.ID, 10),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        user.Role.String(),
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}, nil
	}
}

func getDisplayName(email string) string {
	return strings.Split(email, "@")[0]
}
