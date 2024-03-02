package service

import (
	"BrainBlitz.com/game/internal/core/dto"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"errors"
	"testing"
)

type mockUserRepository struct{}
type mockDuplicateUserRepository struct{}
type mockInvalidUserRepository struct{}

func (m *mockUserRepository) InsertUser(dto dto.UserDTO) error {
	if dto.Username == "test_user" {
		return repository.DuplicateUser
	}
	return nil
}

func (m *mockDuplicateUserRepository) InsertUser(dto dto.UserDTO) error {
	return repository.DuplicateUser
}

func (m *mockInvalidUserRepository) InsertUser(dto dto.UserDTO) error {
	return errors.New("")
}

func TestUserService_SignUp_Success(t *testing.T) {
	userService := NewUserService(&mockUserRepository{})
	req := request.SignUpRequest{
		Email:    "testasf@gmail.com",
		Password: "12345",
	}
	res := userService.SignUp(&req)
	if !res.Status {
		t.Errorf("Expected status to be true, got false")
	}
	data := res.Data.(response.SignUpDataResponse)
	if data.DisplayName == "" {
		t.Errorf("expected non-empty display name, got empty")
	}
}

func TestUserService_SignUp_InvalidUsername(t *testing.T) {
	userRepo := &mockUserRepository{}
	userService := NewUserService(userRepo)

	req := &request.SignUpRequest{
		Email:    "",
		Password: "12345",
	}

	res := userService.SignUp(req)
	if res.Status {
		t.Errorf("expected status to be false, got true")
	}
	if res.ErrorCode != error_code.BadRequest {
		t.Errorf("expected error code to be BadRequest, got %s", res.ErrorCode)
	}
}

func TestUserService_SignUp_InvalidPassword(t *testing.T) {
	userRepo := &mockUserRepository{}
	userService := NewUserService(userRepo)

	req := &request.SignUpRequest{
		Email:    "asdflkjasfd@gmail.com",
		Password: "",
	}

	res := userService.SignUp(req)
	if res.Status {
		t.Errorf("expected status to be false, got true")
	}
	if res.ErrorCode != error_code.BadRequest {
		t.Errorf("expected error code to be BadRequest, got %s", res.ErrorCode)
	}
	if res.ErrorMessage != invalidPasswordErrMsg {
		t.Errorf("expected error message to be %s, got %s", invalidPasswordErrMsg, res.ErrorCode)
	}
}

func TestUserService_SignUp_DuplicateUser(t *testing.T) {
	userRepo := &mockDuplicateUserRepository{}
	userService := NewUserService(userRepo)

	req := &request.SignUpRequest{
		Email:    "asdflkjasfd@gmail.com",
		Password: "fasf",
	}

	res := userService.SignUp(req)

	if res.Status {
		t.Errorf("expected status to be false, got true")
	}
	if res.ErrorCode != error_code.DuplicateUser {
		t.Errorf("expected error code to be BadRequest, got %s", res.ErrorCode)
	}
}

func TestUserService_SignUp_InternalError(t *testing.T) {
	userRepo := &mockInvalidUserRepository{}
	userService := NewUserService(userRepo)

	req := &request.SignUpRequest{
		Email:    "asdflkjasfd@gmail.com",
		Password: "fasf",
	}

	res := userService.SignUp(req)

	if res.Status {
		t.Errorf("expected status to be false, got true")
	}
	if res.ErrorCode == error_code.DuplicateUser {
		t.Errorf("expected error code to be BadRequest, got %s", res.ErrorCode)
	}

	if res.ErrorMessage != error_code.InternalErrMsg {
		t.Errorf("expected error message to be %s, got %s", error_code.InternalErrMsg, res.ErrorMessage)
	}
}
