package service

import (
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/dto"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/email"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"strings"
)

const (
	invalidPasswordErrMsg = "invalid password"
)

type UserService struct {
	userRepo    repository.UserRepository
	authService service.AuthGenerator
}

func NewUserService(userRepo repository.UserRepository, authService service.AuthGenerator) service.UserService {
	return &UserService{
		userRepo:    userRepo,
		authService: authService,
	}
}

func (us UserService) SignUp(request *request.SignUpRequest) (response.SignUpResponse, error) {
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

	userDto := dto.UserDTO{
		Username:       request.Email,
		HashedPassword: hashPassword,
		DisplayName:    getDisplayName(request.Email),
		CreatedAt:      currentTime,
		UpdatedAt:      currentTime,
	}

	//save a new user
	err = us.userRepo.InsertUser(userDto)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return response.SignUpResponse{}, richerror.New(op).
				WithError(err).
				WithKind(richerror.KindInvalid).
				WithMessage(errmsg.DuplicateUsername)
		}
		fmt.Errorf("errorIn: %v", err)
		return response.SignUpResponse{}, richerror.New(op).
			WithError(err).
			WithKind(richerror.KindUnexpected)
	}

	// create data response
	return response.SignUpResponse{
		DisplayName: userDto.DisplayName,
	}, nil
}

func getDisplayName(email string) string {
	return strings.Split(email, "@")[0]
}

func (us UserService) createFailedResponse(errorCode int, message string) *response.Response {
	return &response.Response{
		Status:       false,
		ErrorCode:    errorCode,
		ErrorMessage: message,
	}
}

func (us UserService) createSuccessResponse(data response.SignUpResponse) *response.Response {
	return &response.Response{
		Data:         data,
		Status:       true,
		ErrorCode:    error_code.Success,
		ErrorMessage: error_code.SuccessErrMsg,
	}
}
