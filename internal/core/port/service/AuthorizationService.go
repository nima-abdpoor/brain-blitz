package service

import entity "BrainBlitz.com/game/entity/auth"

type AuthorizationService interface {
	HasAccess(role entity.Role, permissions ...entity.PermissionTitle) (bool, error)
}
