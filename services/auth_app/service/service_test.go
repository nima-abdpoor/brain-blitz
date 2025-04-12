package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// no op
}

func setupTestService() Service {
	return NewService(Config{
		SecretKey:              "test-secret",
		AccessTokenExpireTime:  1 * time.Hour,
		RefreshTokenExpireTime: 24 * time.Hour,
	}, MockLogger{})
}

func TestCreateAccessToken(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	t.Run("valid access token creation", func(t *testing.T) {
		req := CreateAccessTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "id", Value: "123"},
				{Key: "role", Value: "admin"},
			},
		}

		res, err := svc.CreateAccessToken(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.AccessToken)
		assert.True(t, res.ExpireTime > 0)
	})

	t.Run("invalid access token request", func(t *testing.T) {
		req := CreateAccessTokenRequest{} // Missing data
		_, err := svc.CreateAccessToken(ctx, req)
		assert.Error(t, err)
	})
}

func TestCreateRefreshToken(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	t.Run("valid refresh token creation", func(t *testing.T) {
		req := CreateRefreshTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "id", Value: "123"},
				{Key: "role", Value: "user"},
			},
		}

		res, err := svc.CreateRefreshToken(ctx, req)
		assert.NoError(t, err)
		assert.NotEmpty(t, res.RefreshToken)
		assert.True(t, res.ExpireTime > 0)
	})

	t.Run("invalid refresh token request", func(t *testing.T) {
		req := CreateRefreshTokenRequest{}
		_, err := svc.CreateRefreshToken(ctx, req)
		assert.Error(t, err)
	})
}

func TestValidateToken(t *testing.T) {
	svc := setupTestService()
	ctx := context.Background()

	// First create a valid token
	createReq := CreateAccessTokenRequest{
		Data: []CreateTokenRequest{
			{Key: "id", Value: "123"},
			{Key: "role", Value: "admin"},
		},
	}
	accessTokenRes, err := svc.CreateAccessToken(ctx, createReq)
	assert.NoError(t, err)

	t.Run("valid token validation", func(t *testing.T) {
		validateReq := ValidateTokenRequest{
			Token: accessTokenRes.AccessToken,
			Data:  []string{"id", "role"},
		}
		res, err := svc.ValidateToken(ctx, validateReq)
		assert.NoError(t, err)
		assert.True(t, res.Valid)
		assert.Len(t, res.AdditionalData, 2)
	})

	t.Run("invalid token", func(t *testing.T) {
		validateReq := ValidateTokenRequest{
			Token: "invalid.token.string",
			Data:  []string{"id"},
		}
		res, err := svc.ValidateToken(ctx, validateReq)
		assert.Error(t, err)
		assert.False(t, res.Valid)
	})

	t.Run("empty token", func(t *testing.T) {
		validateReq := ValidateTokenRequest{
			Token: "",
			Data:  []string{"id"},
		}
		_, err := svc.ValidateToken(ctx, validateReq)
		assert.Error(t, err)
	})
}
