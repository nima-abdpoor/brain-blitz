package service

import (
	entity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/port/repository"
	"BrainBlitz.com/game/internal/core/port/service"
	"BrainBlitz.com/game/pkg/slice"
)

type Authorization struct {
	repo repository.AuthorizationRepository
}

func NewAuthorizationService(repo repository.AuthorizationRepository) service.AuthorizationService {
	return &Authorization{
		repo: repo,
	}
}

func (a Authorization) HasAccess(role entity.Role, permissions ...entity.PermissionTitle) (bool, error) {
	const op = "service.Authorization.HasAccess"
	if userPermissions, err := a.repo.GetUserPermissionTitles(role); err != nil {
		return false, err
	} else {
		for _, requiredPermission := range permissions {
			if !slice.IsExists(string(requiredPermission), userPermissions) {
				//todo check if logger or metrics needed
				//log.Println(fmt.Sprintf("doesnt have %s permission", requiredPermission))
				return false, nil
			}
		}
		return true, nil
	}
}
