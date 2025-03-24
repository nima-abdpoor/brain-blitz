package repository

import entity "BrainBlitz.com/game/entity/auth"

type AuthorizationRepository interface {
	GetUserPermissionTitles(role entity.Role) ([]entity.PermissionTitle, error)
}
