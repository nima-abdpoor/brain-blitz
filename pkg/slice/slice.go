package slice

import entity "BrainBlitz.com/game/entity/auth"

func IsExists(msg string, strings []entity.PermissionTitle) bool {
	for _, str := range strings {
		if string(str) == msg {
			return true
		}
	}
	return false
}
