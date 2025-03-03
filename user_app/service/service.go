package service

import (
	authEntity "BrainBlitz.com/game/entity/auth"
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/logger"
	cachemanager "BrainBlitz.com/game/pkg/cache_manager"
	"BrainBlitz.com/game/pkg/email"
	"BrainBlitz.com/game/pkg/errmsg"
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

func (us Service) SignUp(request *request.SignUpRequest) (response.SignUpResponse, error) {
	const op = "service.SignUp"
	if !email.IsValid(request.Email) {
		return response.SignUpResponse{}, richerror.New(op).
			WithMeta(map[string]interface{}{"email": request.Email}).
			WithMessage(errmsg.InvalidUserNameErrMsg)
	}

	if len(request.Password) == 0 {
		return response.SignUpResponse{}, richerror.New(op).
			WithMessage(errmsg.InvalidPasswordErrMsg).
			WithMeta(map[string]interface{}{"password": request.Password})
	}

	currentTime := utils.GetUTCCurrentMillis()

	hashPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return response.SignUpResponse{}, richerror.New(op).
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

	//save a new user
	err = us.repository.InsertUser(userDto)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return response.SignUpResponse{}, richerror.New(op).
				WithError(err).
				WithKind(richerror.KindInvalid).
				WithMessage(errmsg.DuplicateUsername)
		}
		//todo add to metrics
		logger.Logger.Named(op).Error("Error in inserting User", zap.String("userDto", fmt.Sprint(userDto)), zap.Error(err))
		return response.SignUpResponse{}, richerror.New(op).
			WithError(err).
			WithKind(richerror.KindUnexpected)
	}

	// create data response
	return response.SignUpResponse{
		DisplayName: userDto.DisplayName,
	}, nil
}

func (us Service) Profile(id int64) (response.ProfileResponse, error) {
	const op = "service.Profile"
	if user, err := us.repository.GetUserById(id); err != nil {
		// todo check if logger needed
		// todo add metrics
		return response.ProfileResponse{}, richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		return response.ProfileResponse{
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
