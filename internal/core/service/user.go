package service

import (
	authEntity "BrainBlitz.com/game/entity/auth"
	entity "BrainBlitz.com/game/entity/user"
	utils "BrainBlitz.com/game/internal/core/common"
	"BrainBlitz.com/game/internal/core/entity/error_code"
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/email"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"fmt"
	"strconv"
	"strings"
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

	userDto := entity.User{
		Username:       request.Email,
		HashedPassword: hashPassword,
		DisplayName:    getDisplayName(request.Email),
		Role:           authEntity.UserRole,
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
		fmt.Println(op, err)
		return response.SignUpResponse{}, richerror.New(op).
			WithError(err).
			WithKind(richerror.KindUnexpected)
	}

	// create data response
	return response.SignUpResponse{
		DisplayName: userDto.DisplayName,
	}, nil
}

func (us UserService) Profile(id int64) (response.ProfileResponse, error) {
	const op = "service.Profile"
	if user, err := us.userRepo.GetUserById(id); err != nil {
		fmt.Println(err)
		_ = fmt.Errorf("error In Getting User: %v", err)
		return response.ProfileResponse{}, richerror.New(op).WithError(err).WithKind(richerror.KindUnexpected)
	} else {
		return response.ProfileResponse{
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
