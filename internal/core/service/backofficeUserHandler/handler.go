package backofficeUserHandler

import (
	"BrainBlitz.com/game/internal/core/model/request"
	"BrainBlitz.com/game/internal/core/model/response"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"strconv"
)

type Service struct {
	repo repository.BackofficeUserRepository
}

func New(repo repository.BackofficeUserRepository) service.BackofficeUserService {
	return Service{
		repo: repo,
	}
}

func (service Service) ListUsers(request *request.ListUserRequest) (response.ListUserResponse, error) {
	const op = "backofficeUserHandler.ListUsers"

	users := []response.User{}
	// todo get offset and limit from input
	if result, err := service.repo.ListUsers(); err != nil {
		//logger.Logger.Named(op).Error("error in listingUsers", zap.Error(err))
		//return response.ListUserResponse{}, richerror.New(op).WithKind(richerror.KindUnexpected).WithError(err)
		return response.ListUserResponse{}, err
	} else {
		for _, user := range result {
			users = append(users, response.User{
				ID:          strconv.FormatInt(user.ID, 10),
				Username:    user.Username,
				DisplayName: user.DisplayName,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
				Role:        user.Role.String(),
			})
		}
		return response.ListUserResponse{
			Users: users,
		}, nil
	}
}
