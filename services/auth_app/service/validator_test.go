package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateCreateAccessTokenRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := CreateAccessTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "user_id", Value: "123"},
			},
		}
		err := ValidateCreateAccessTokenRequest(req)
		assert.NoError(t, err)
	})

	t.Run("missing data", func(t *testing.T) {
		req := CreateAccessTokenRequest{}
		err := ValidateCreateAccessTokenRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrValidationDataRequired)
	})

	t.Run("missing key or value in data", func(t *testing.T) {
		req := CreateAccessTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "", Value: "abc"},
				{Key: "user_id", Value: ""},
			},
		}
		err := ValidateCreateAccessTokenRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrValidationRequired)
	})
}

func TestValidateCreateRefreshTokenRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := CreateRefreshTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "device", Value: "mobile"},
			},
		}
		err := ValidateCreateRefreshTokenRequest(req)
		assert.NoError(t, err)
	})

	t.Run("missing data", func(t *testing.T) {
		req := CreateRefreshTokenRequest{}
		err := ValidateCreateRefreshTokenRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrValidationDataRequired)
	})

	t.Run("invalid data item", func(t *testing.T) {
		req := CreateRefreshTokenRequest{
			Data: []CreateTokenRequest{
				{Key: "", Value: "android"},
			},
		}
		err := ValidateCreateRefreshTokenRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrValidationRequired)
	})
}

func TestValidateValidateTokenRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := ValidateTokenRequest{
			Token: "some-token",
			Data:  []string{"user_id", "device"},
		}
		err := ValidateValidateTokenRequest(req)
		assert.NoError(t, err)
	})

	t.Run("missing token", func(t *testing.T) {
		req := ValidateTokenRequest{
			Data: []string{"user_id"},
		}
		err := ValidateValidateTokenRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrValidationDataRequired)
	})
}
