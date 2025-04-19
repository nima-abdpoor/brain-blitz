package http

import (
	auth_adapter "BrainBlitz.com/game/adapter/auth"
	cachemanager "BrainBlitz.com/game/pkg/cache_manager"
	utils "BrainBlitz.com/game/pkg/common"
	errApp "BrainBlitz.com/game/pkg/err_app"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"BrainBlitz.com/game/services/user_app/service"
	"bytes"
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
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

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) InsertUser(ctx context.Context, user service.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *mockRepository) GetUser(ctx context.Context, email string) (service.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(service.User), args.Error(1)
}

func (m *mockRepository) GetUserById(ctx context.Context, id string) (service.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(service.User), args.Error(1)
}

type mockTokenClient struct {
	mock.Mock
}

func (m *mockTokenClient) GetAccessToken(ctx context.Context, req auth_adapter.CreateAccessTokenRequest) (auth_adapter.CreateAccessTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateAccessTokenResponse), args.Error(1)
}

func (m *mockTokenClient) GetRefreshToken(ctx context.Context, req auth_adapter.CreateRefreshTokenRequest) (auth_adapter.CreateRefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateRefreshTokenResponse), args.Error(1)
}

func TestLoginIntegration(t *testing.T) {
	e := echo.New()

	mockRepo := new(mockRepository)
	mockClient := new(mockTokenClient)
	mockCache := cachemanager.NewCacheManager(nil)

	svc := service.NewService(mockRepo, *mockCache, mockClient, MockLogger{})

	h := NewHandler(svc, MockLogger{})

	e.POST("/login", h.Login)

	t.Run("error-invalid-password", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("password123")
		testUser := service.User{
			ID:             1,
			Username:       "testUser@example.com",
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			CreatedAt:      uint64(time.Now().UnixMilli()),
			UpdatedAt:      uint64(time.Now().UnixMilli()),
			Role:           service.UserRole,
		}

		mockRepo.On("GetUser", mock.Anything, "testUser@example.com").Return(testUser, nil)

		loginReq := service.LoginRequest{
			Email:    "testUser@example.com",
			Password: "InCorrectPassword",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.Login(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)

			var res errApp.HTTPErrMessage
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, errmsg.InvalidUserNameOrPasswordErrMsg, res.Message)
		}
	})

	t.Run("success", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("password123")
		testUser := service.User{
			ID:             1,
			Username:       "testUser@example.com",
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			CreatedAt:      uint64(time.Now().UnixMilli()),
			UpdatedAt:      uint64(time.Now().UnixMilli()),
			Role:           service.UserRole,
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, "testUser@example.com").Return(testUser, nil)

		mockClient.On("GetAccessToken", mock.Anything, mock.Anything).
			Return(auth_adapter.CreateAccessTokenResponse{
				AccessToken: "test-access-token",
				ExpireTime:  3600,
			}, nil)

		mockClient.On("GetRefreshToken", mock.Anything, mock.Anything).
			Return(auth_adapter.CreateRefreshTokenResponse{
				RefreshToken: "test-refresh-token",
				ExpireTime:   7200,
			}, nil)

		loginReq := service.LoginRequest{
			Email:    "testUser@example.com",
			Password: "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, h.Login(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			var res service.LoginResponse
			err := json.Unmarshal(rec.Body.Bytes(), &res)
			assert.NoError(t, err)

			assert.Equal(t, "1", res.ID)
			assert.Equal(t, "test-access-token", res.AccessToken)
			assert.Equal(t, "test-refresh-token", res.RefreshToken)
		}
	})
}
