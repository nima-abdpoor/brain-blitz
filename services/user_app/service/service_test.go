package service

import (
	"BrainBlitz.com/game/adapter/auth"
	"BrainBlitz.com/game/pkg/cache_manager"
	"BrainBlitz.com/game/pkg/common"
	errmsg "BrainBlitz.com/game/pkg/err_msg"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockRepository struct {
	mock.Mock
}

type MockLogger struct{}

func (m MockLogger) Info(msg string, args ...any)  {}
func (m MockLogger) Warn(msg string, args ...any)  {}
func (m MockLogger) Debug(msg string, args ...any) {}
func (m MockLogger) Error(msg string, keysAndValues ...interface{}) {
	// no op
}

func (m *MockRepository) InsertUser(ctx context.Context, user User) (int, error) {
	args := m.Called(ctx, user)
	return 1, args.Error(1)
}

func (m *MockRepository) GetUser(ctx context.Context, email string) (User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockRepository) GetUserById(ctx context.Context, id string) (User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

type MockGRPCClient struct {
	mock.Mock
}

type MockTokenClient struct {
	mock.Mock
}

func (m *MockTokenClient) GetAccessToken(ctx context.Context, req auth_adapter.CreateAccessTokenRequest) (auth_adapter.CreateAccessTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateAccessTokenResponse), args.Error(1)
}

func (m *MockTokenClient) GetRefreshToken(ctx context.Context, req auth_adapter.CreateRefreshTokenRequest) (auth_adapter.CreateRefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateRefreshTokenResponse), args.Error(1)
}

func (m *MockGRPCClient) GetAccessToken(ctx context.Context, req auth_adapter.CreateAccessTokenRequest) (auth_adapter.CreateAccessTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateAccessTokenResponse), args.Error(1)
}

func (m *MockGRPCClient) GetRefreshToken(ctx context.Context, req auth_adapter.CreateRefreshTokenRequest) (auth_adapter.CreateRefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(auth_adapter.CreateRefreshTokenResponse), args.Error(1)
}

func TestService_SignUp(t *testing.T) {
	mockRepo := new(MockRepository)
	mockClient := &MockTokenClient{}
	mockLogger := MockLogger{}
	mockCache := cachemanager.NewCacheManager(nil)
	service := NewService(mockRepo, *mockCache, mockClient, mockLogger)

	t.Run("success", func(t *testing.T) {
		req := SignUpRequest{
			Email:    "testUser@example.com",
			Password: "securepassword",
		}

		mockRepo.On("InsertUser", mock.Anything, mock.Anything).Return(int64(1), nil)

		resp, err := service.SignUp(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "testUser", resp.DisplayName)
	})

	t.Run("error-invalid-email", func(t *testing.T) {
		req := SignUpRequest{
			Email:    "test@example.com",
			Password: "securepassword",
		}

		resp, err := service.SignUp(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, errmsg.InvalidInputErrMsg, err.Error())
		assert.Equal(t, "", resp.DisplayName)
	})

	t.Run("error-invalid-password", func(t *testing.T) {
		req := SignUpRequest{
			Email: "testUser@example.com",
		}

		resp, err := service.SignUp(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, errmsg.InvalidUserNameOrPasswordErrMsg, err.Error())
		assert.Equal(t, "", resp.DisplayName)
	})

	t.Run("error-duplicate-user", func(t *testing.T) {
		req := SignUpRequest{
			Email:    "testUser@example.com",
			Password: "securepassword",
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("InsertUser", mock.Anything, mock.Anything).Return(int64(2), fmt.Errorf("duplicate user"))

		resp, err := service.SignUp(context.Background(), req)
		assert.Equal(t, errmsg.DuplicateUsername, err.Error())
		assert.Equal(t, "", resp.DisplayName)
	})

	t.Run("error-database", func(t *testing.T) {
		req := SignUpRequest{
			Email:    "testUser@example.com",
			Password: "securepassword",
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("InsertUser", mock.Anything, mock.Anything).Return(int64(1), fmt.Errorf("unkonwn error"))

		resp, err := service.SignUp(context.Background(), req)
		assert.Error(t, err)
		assert.Equal(t, errmsg.SomeThingWentWrong, err.Error())
		assert.Equal(t, "", resp.DisplayName)
	})
}

func TestService_Login(t *testing.T) {
	mockRepo := new(MockRepository)
	mockClient := &MockGRPCClient{}
	mockLogger := MockLogger{}
	mockCache := cachemanager.NewCacheManager(nil)
	service := NewService(mockRepo, *mockCache, mockClient, mockLogger)

	t.Run("error-invalid-email", func(t *testing.T) {
		email := "test@example.com"

		req := LoginRequest{Email: email}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.InvalidUserNameOrPasswordErrMsg, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("error-invalid-password", func(t *testing.T) {
		email := "testUser@example.com"

		req := LoginRequest{Email: email}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.InvalidUserNameOrPasswordErrMsg, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("error-database", func(t *testing.T) {
		email := "testUser@example.com"
		password := "securepassword"

		mockUser := User{}
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, email).Return(mockUser, fmt.Errorf("database error"))

		req := LoginRequest{Email: email, Password: password}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.SomeThingWentWrong, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("error-wrong-password", func(t *testing.T) {
		email := "testUser@example.com"
		password := "securepassword"
		hashedPassword := "another password"

		mockUser := User{
			ID:             1,
			Username:       email,
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			Role:           UserRole,
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, email).Return(mockUser, nil)

		req := LoginRequest{Email: email, Password: password}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.InvalidUserNameOrPasswordErrMsg, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("error-access-token", func(t *testing.T) {
		email := "testUser@example.com"
		password := "securepassword"
		hashedPassword, _ := utils.HashPassword(password)

		mockUser := User{
			ID:             1,
			Username:       email,
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			Role:           UserRole,
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, email).Return(mockUser, nil)

		mockClient.On("GetAccessToken", mock.Anything, mock.Anything).Return(auth_adapter.CreateAccessTokenResponse{
			AccessToken: "access-token",
		}, fmt.Errorf("service error"))

		req := LoginRequest{Email: email, Password: password}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.SomeThingWentWrong, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("error-refresh-token", func(t *testing.T) {
		email := "testUser@example.com"
		password := "securepassword"
		hashedPassword, _ := utils.HashPassword(password)

		mockUser := User{
			ID:             1,
			Username:       email,
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			Role:           UserRole,
		}
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, email).Return(mockUser, nil)

		mockClient.ExpectedCalls = nil
		mockClient.On("GetAccessToken", mock.Anything, mock.Anything).Return(auth_adapter.CreateAccessTokenResponse{
			AccessToken: "access-token",
		}, nil)
		mockClient.On("GetRefreshToken", mock.Anything, mock.Anything).Return(auth_adapter.CreateRefreshTokenResponse{
			RefreshToken: "refresh-token",
		}, fmt.Errorf("service error"))

		req := LoginRequest{Email: email, Password: password}
		resp, err := service.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, errmsg.SomeThingWentWrong, err.Error())
		assert.Equal(t, "", resp.AccessToken)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.RefreshToken)
	})

	t.Run("success", func(t *testing.T) {
		email := "testUser@example.com"
		password := "pass123"
		hashedPassword, _ := utils.HashPassword(password)

		mockUser := User{
			ID:             1,
			Username:       email,
			HashedPassword: hashedPassword,
			DisplayName:    "testUser",
			Role:           UserRole,
		}

		mockRepo.ExpectedCalls = nil
		mockClient.ExpectedCalls = nil
		mockRepo.On("GetUser", mock.Anything, email).Return(mockUser, nil)
		mockClient.On("GetAccessToken", mock.Anything, mock.Anything).Return(auth_adapter.CreateAccessTokenResponse{
			AccessToken: "access-token",
		}, nil)
		mockClient.On("GetRefreshToken", mock.Anything, mock.Anything).Return(auth_adapter.CreateRefreshTokenResponse{
			RefreshToken: "refresh-token",
		}, nil)

		req := LoginRequest{Email: email, Password: password}
		resp, err := service.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "access-token", resp.AccessToken)
		assert.Equal(t, "refresh-token", resp.RefreshToken)
	})
}

func TestService_Profile(t *testing.T) {
	mockRepo := new(MockRepository)
	mockClient := new(MockGRPCClient)
	mockLogger := MockLogger{}
	mockCache := cachemanager.NewCacheManager(nil)
	service := NewService(mockRepo, *mockCache, mockClient, mockLogger)

	t.Run("error-invalid-user", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUserById", mock.Anything, "1").Return(User{}, fmt.Errorf("record not found"))

		resp, err := service.Profile(context.Background(), ProfileRequest{ID: "1"})
		fmt.Println(resp)
		assert.Error(t, err)
		assert.Equal(t, errmsg.UserNotFoundErrMsg, err.Error())
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.Role)
		assert.Equal(t, uint64(0), resp.CreatedAt)
		assert.Equal(t, uint64(0), resp.UpdatedAt)
		assert.Equal(t, "", resp.Username)
		assert.Equal(t, "", resp.DisplayName)
	})

	t.Run("error-database", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUserById", mock.Anything, "1").Return(User{}, fmt.Errorf("database error"))

		resp, err := service.Profile(context.Background(), ProfileRequest{ID: "1"})
		fmt.Println(resp)
		assert.Error(t, err)
		assert.Equal(t, errmsg.SomeThingWentWrong, err.Error())
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, "", resp.Role)
		assert.Equal(t, uint64(0), resp.CreatedAt)
		assert.Equal(t, uint64(0), resp.UpdatedAt)
		assert.Equal(t, "", resp.Username)
		assert.Equal(t, "", resp.DisplayName)
	})

	t.Run("success", func(t *testing.T) {
		user := User{
			ID:          1,
			Username:    "testUser@example.com",
			DisplayName: "testUser",
			Role:        UserRole,
			CreatedAt:   uint64(time.Now().UnixMilli()),
			UpdatedAt:   uint64(time.Now().UnixMilli()),
		}

		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetUserById", mock.Anything, "1").Return(user, nil)

		resp, err := service.Profile(context.Background(), ProfileRequest{ID: "1"})
		assert.NoError(t, err)
		assert.Equal(t, "1", resp.ID)
		assert.Equal(t, "testUser@example.com", resp.Username)
		assert.Equal(t, "testUser", resp.DisplayName)
	})
}
