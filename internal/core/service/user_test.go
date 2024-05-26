package service

import (
	entity "BrainBlitz.com/game/entity/user"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"testing"
)

type mockUserRepository struct{}
type mockAuthGenerator struct{}
type mockInvalidUserRepository struct{}

func (m mockAuthGenerator) CreateAccessToken(data map[string]string) (string, error) {
	return "AccessToken", nil
}

func (m mockAuthGenerator) CreateRefreshToken(data map[string]string) (string, error) {
	return "RefreshToken", nil
}

func (m mockAuthGenerator) ValidateToken(data []string, token string) (bool, map[string]interface{}, error) {
	return true, nil, nil
}

func (m *mockUserRepository) InsertUser(user entity.User) error {
	return nil
}

func (m *mockUserRepository) GetUser(email string) (entity.User, error) {
	return entity.User{}, nil
}

func (m *mockUserRepository) GetUserById(id int64) (entity.User, error) {
	return entity.User{}, nil
}

func (m *mockInvalidUserRepository) InsertUser(user entity.User) error {
	return richerror.New("service.test_GetUser").WithKind(richerror.KindUnexpected).WithMessage(errmsg.SomeThingWentWrong)
}
func (m *mockInvalidUserRepository) GetUser(email string) (entity.User, error) {
	return entity.User{}, richerror.New("service.test_GetUser").WithKind(richerror.KindUnexpected)
}
func (m *mockInvalidUserRepository) GetUserById(id int64) (entity.User, error) {
	return entity.User{}, richerror.New("service.test_GetUser").WithKind(richerror.KindUnexpected).WithMessage(errmsg.SomeThingWentWrong)
}

// signUp tests
func TestUserService_SignUp_Success(t *testing.T) {
	userRepo := &mockUserRepository{}
	authGenerator := &mockAuthGenerator{}
	userService := NewUserService(userRepo, authGenerator)
	req := request.SignUpRequest{
		Email:    "testasf@gmail.com",
		Password: "12345",
	}
	res, err := userService.SignUp(&req)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
		return
	}
	if res.DisplayName == "" {
		t.Errorf("expected non-empty display name, got empty")
	}
	if res.DisplayName != "testasf" {
		t.Errorf("display name is wrong!")
	}
}

func TestUserService_SignUp_InvalidUsername(t *testing.T) {
	userRepo := &mockUserRepository{}
	authGenerator := &mockAuthGenerator{}

	userService := NewUserService(userRepo, authGenerator)

	req := &request.SignUpRequest{
		Email:    "",
		Password: "12345",
	}

	_, err := userService.SignUp(req)
	if err == nil {
		t.Errorf("expected error to be nil, got %v", err)
		return
	}
	if err.Error() != errmsg.InvalidUserNameErrMsg {
		t.Errorf("expected error to be %v, got %v", errmsg.InvalidUserNameErrMsg, err)
	}
}

func TestUserService_SignUp_InvalidPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	authGenerator := &mockAuthGenerator{}

	userService := NewUserService(userRepo, authGenerator)

	req := &request.SignUpRequest{
		Email:    "asdflkjasfd@gmail.com",
		Password: "",
	}

	_, err := userService.SignUp(req)
	if err == nil {
		t.Errorf("expected error not to be nil, got %v", err)
		return
	}
	if err.Error() != errmsg.InvalidPasswordErrMsg {
		t.Errorf("expected error to be %v, got %v", errmsg.InvalidPasswordErrMsg, err)
		return
	}
}

func TestUserService_SignUp_InternalError(t *testing.T) {
	userRepo := &mockInvalidUserRepository{}
	authGenerator := &mockAuthGenerator{}

	userService := NewUserService(userRepo, authGenerator)

	req := &request.SignUpRequest{
		Email:    "asdflkjasfd@gmail.com",
		Password: "fasf",
	}

	_, err := userService.SignUp(req)

	if err == nil {
		t.Errorf("expected error not to be nil, got %v", err)
		return
	}
}
