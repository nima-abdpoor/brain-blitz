package service

import (
	"BrainBlitz.com/game/adapter/auth"
	authEntity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/logger"
	cachemanager "BrainBlitz.com/game/pkg/cache_manager"
	utils2 "BrainBlitz.com/game/pkg/common"
	"BrainBlitz.com/game/pkg/email"
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"net/http"
	"strconv"
	"strings"
)

type Repository interface {
	InsertUser(ctx context.Context, user User) (int, error)
	GetUser(ctx context.Context, email string) (User, error)
	GetUserById(ctx context.Context, id string) (User, error)
}

type Service struct {
	repository   Repository
	grpcClient   *auth_adapter.Client
	CacheManager cachemanager.CacheManager
}

func NewService(repository Repository, cm cachemanager.CacheManager, grpcClient *auth_adapter.Client) Service {
	return Service{
		repository:   repository,
		CacheManager: cm,
		grpcClient:   grpcClient,
	}
}

func (s Service) SignUp(ctx context.Context, request SignUpRequest) (SignUpResponse, error) {
	const op = "service.SignUp"
	if !email.IsValid(request.Email) {
		return SignUpResponse{}, errApp.Wrap(op, nil, errApp.ErrInvalidInput, map[string]string{
			"message": "InvalidUserNameErrMsg",
			"data":    fmt.Sprint(request),
		})
	}

	if len(request.Password) == 0 {
		return SignUpResponse{}, errApp.Wrap(op, nil, errApp.ErrInvalidLOGIN, map[string]string{
			"message": "InvalidPasswordErrMsg",
			"data":    fmt.Sprint(request),
		})
	}

	currentTime := utils2.GetUTCCurrentMillis()

	hashPassword, err := utils2.HashPassword(request.Password)
	if err != nil {
		return SignUpResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
			"message": "BcryptErrorHashingPassword",
			"data":    fmt.Sprint(request),
		})
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
		if strings.Contains(err.Error(), "duplicate") {
			return SignUpResponse{}, errApp.New(op, "DUPLICATE_USERNAME", errmsg.DuplicateUsername, http.StatusBadRequest, codes.InvalidArgument, map[string]string{
				"message": "Error in inserting User",
				"data":    fmt.Sprint(userDto),
			})
		}
		//todo add to metrics
		return SignUpResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
			"message": "Error in inserting User",
			"data":    fmt.Sprint(userDto),
		})
	}

	return SignUpResponse{
		DisplayName: userDto.DisplayName,
	}, nil
}

func (s Service) Login(ctx context.Context, request LoginRequest) (LoginResponse, error) {
	const op = "service.Login"
	if !email.IsValid(request.Email) {
		return LoginResponse{}, errApp.Wrap(op, nil, errApp.ErrInvalidLOGIN, map[string]string{
			"message": "invalid Email",
			"data":    fmt.Sprint(request),
		})
	}

	if len(request.Password) == 0 {
		return LoginResponse{}, errApp.Wrap(op, nil, errApp.ErrInvalidLOGIN, map[string]string{
			"message": "invalid Password",
			"data":    fmt.Sprint(request),
		})
	}

	if user, err := s.repository.GetUser(ctx, request.Email); err != nil {
		logger.Logger.Named(op).Error("error In Getting User", zap.String("email", request.Email), zap.Error(err))
		return LoginResponse{}, err
	} else {
		result := utils2.CheckPasswordHash(request.Password, user.HashedPassword)
		if result {
			data := make([]auth_adapter.CreateTokenRequest, 0)
			data = append(data, auth_adapter.CreateTokenRequest{
				Key:   "user",
				Value: strconv.FormatInt(user.ID, 10),
			})
			data = append(data, auth_adapter.CreateTokenRequest{
				Key:   "role",
				Value: user.Role.String(),
			})

			accessTokenResponse, err := s.grpcClient.GetAccessToken(ctx, auth_adapter.CreateAccessTokenRequest{
				Data: data,
			})
			if err != nil {
				// todo add metrics
				return LoginResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
					"message": "error creating Access Token",
					"data":    fmt.Sprint(data),
				})
			}
			refreshTokenResponse, err := s.grpcClient.GetRefreshToken(ctx, auth_adapter.CreateRefreshTokenRequest{
				Data: data,
			})
			if err != nil {
				return LoginResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, map[string]string{
					"message": "error In Creating Refresh Token",
					"data":    fmt.Sprint(data),
				})
			}
			return LoginResponse{
				ID:           strconv.FormatInt(user.ID, 10),
				AccessToken:  accessTokenResponse.AccessToken,
				RefreshToken: refreshTokenResponse.RefreshToken,
			}, nil
		} else {
			return LoginResponse{}, errApp.Wrap(op, err, errApp.ErrInvalidLOGIN, map[string]string{
				"request": fmt.Sprint(request),
			})
		}
	}
}

func (s Service) Profile(ctx context.Context, request ProfileRequest) (ProfileResponse, error) {
	const op = "service.Profile"
	if user, err := s.repository.GetUserById(ctx, request.ID); err != nil {
		// todo check if logger needed
		// todo add metrics
		return ProfileResponse{}, errApp.Wrap(op, err, errApp.ErrInternal, nil)
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
