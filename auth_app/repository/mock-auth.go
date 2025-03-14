package repository

import (
	"BrainBlitz.com/game/auth_app/service"
	"context"
)

type MockAuthGenerator struct {
	MockedValidateToken      func(ctx context.Context, request service.ValidateTokenRequest) (service.ValidateTokenResponse, error)
	MockedCreateAccessToken  func(ctx context.Context, request service.CreateAccessTokenRequest) (service.CreateAccessTokenResponse, error)
	MockedCreateRefreshToken func(ctx context.Context, request service.CreateRefreshTokenRequest) (service.CreateRefreshTokenResponse, error)
}

func NewMockAuthGenerator(
	mockedValidateToken func(ctx context.Context, request service.ValidateTokenRequest) (service.ValidateTokenResponse, error),
	mockedCreateAccessToken func(ctx context.Context, request service.CreateAccessTokenRequest) (service.CreateAccessTokenResponse, error),
	mockedCreateRefreshToken func(ctx context.Context, request service.CreateRefreshTokenRequest) (service.CreateRefreshTokenResponse, error),
) service.AuthManagement {
	return &MockAuthGenerator{
		MockedValidateToken:      mockedValidateToken,
		MockedCreateAccessToken:  mockedCreateAccessToken,
		MockedCreateRefreshToken: mockedCreateRefreshToken,
	}
}

func (m MockAuthGenerator) CreateAccessToken(ctx context.Context, request service.CreateAccessTokenRequest) (service.CreateAccessTokenResponse, error) {
	return m.MockedCreateAccessToken(ctx, request)
}

func (m MockAuthGenerator) CreateRefreshToken(ctx context.Context, request service.CreateRefreshTokenRequest) (service.CreateRefreshTokenResponse, error) {
	return m.MockedCreateRefreshToken(ctx, request)
}

func (m MockAuthGenerator) ValidateToken(ctx context.Context, request service.ValidateTokenRequest) (service.ValidateTokenResponse, error) {
	return m.MockedValidateToken(ctx, request)
}
