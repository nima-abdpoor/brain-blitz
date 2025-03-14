package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type AuthManagement interface {
	CreateAccessToken(ctx context.Context, request CreateAccessTokenRequest) (CreateAccessTokenResponse, error)
	CreateRefreshToken(ctx context.Context, request CreateRefreshTokenRequest) (CreateRefreshTokenResponse, error)
	ValidateToken(ctx context.Context, request ValidateTokenRequest) (ValidateTokenResponse, error)
}

const (
	expireTimeKey = "exp"
	jwtIdKey      = "jti"
	jwtIssuedAt   = "iat"
)

var additionalData = []string{"id", "role"}

type Config struct {
	SecretKey              string        `koanf:"secret_key"`
	AccessTokenExpireTime  time.Duration `koanf:"access_token_expire_time"`
	RefreshTokenExpireTime time.Duration `koanf:"refresh_token_expire_time"`
}

type Service struct {
	config Config
	logger *slog.Logger
}

func NewService(config Config, logger *slog.Logger) Service {
	return Service{
		config: config,
		logger: logger,
	}
}

func (svc Service) CreateAccessToken(ctx context.Context, request CreateAccessTokenRequest) (CreateAccessTokenResponse, error) {
	err := ValidateCreateAccessTokenRequest(request)
	if err != nil {
		return CreateAccessTokenResponse{}, err
	}

	claims := toJWTClaims(request.Data)
	claims[expireTimeKey] = svc.config.AccessTokenExpireTime
	claims[jwtIdKey] = uuid.NewString()
	claims[jwtIssuedAt] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(svc.config.SecretKey))
	if err != nil {
		return CreateAccessTokenResponse{}, err
	}
	return CreateAccessTokenResponse{
		AccessToken: signedString,
		ExpireTime:  svc.config.AccessTokenExpireTime.Milliseconds(),
	}, nil
}

func (svc Service) CreateRefreshToken(ctx context.Context, request CreateRefreshTokenRequest) (CreateRefreshTokenResponse, error) {
	err := ValidateCreateRefreshTokenRequest(request)
	if err != nil {
		return CreateRefreshTokenResponse{}, err
	}

	claims := toJWTClaims(request.Data)
	claims[expireTimeKey] = svc.config.RefreshTokenExpireTime
	claims[jwtIdKey] = uuid.NewString()
	claims[jwtIssuedAt] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(svc.config.SecretKey))
	if err != nil {
		return CreateRefreshTokenResponse{}, err
	}
	return CreateRefreshTokenResponse{
		RefreshToken: signedString,
		ExpireTime:   svc.config.RefreshTokenExpireTime.Milliseconds(),
	}, nil
}

func (svc Service) ValidateToken(ctx context.Context, request ValidateTokenRequest) (ValidateTokenResponse, error) {
	op := "service.ValidateToken"

	err := ValidateValidateTokenRequest(request)
	if err != nil {
		return ValidateTokenResponse{}, err
	}

	token, err := jwt.Parse(request.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.config.SecretKey), nil
	})

	if err != nil {
		return ValidateTokenResponse{Valid: false}, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); !ok {
		// todo add metrics
		svc.logger.Error(op, "casting Problem with JWT Claims")
		return ValidateTokenResponse{Valid: false}, fmt.Errorf("casting Problem with JWT Claims")
	} else {
		data := make(map[string]struct{})
		for _, str := range additionalData {
			data[str] = struct{}{}
		}
		for _, str := range request.Data {
			data[str] = struct{}{}
		}
		return ValidateTokenResponse{
			Valid:          true,
			AdditionalData: toMapData(data, claims),
		}, nil
	}
}
