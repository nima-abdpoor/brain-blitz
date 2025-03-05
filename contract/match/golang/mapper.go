package match

import (
	entity "BrainBlitz.com/game/entity/game"
	"BrainBlitz.com/game/match_app/service"
)

func MapFromEntityToProtoMessage(matchedUsers []service.MatchedUsers) *AllMatchedUsers {
	finalUsers := make([]*service.MatchedUsers, 0)
	for _, user := range matchedUsers {
		finalUsers = append(finalUsers, &service.MatchedUsers{
			UserId:   user.UserId,
			Category: service.MapFromCategory(user.Category),
		})
	}

	return &AllMatchedUsers{
		Users: finalUsers,
	}
}

func MapToEntityToProtoMessage(users *AllMatchedUsers) []entity.MatchedUsers {
	finalUsers := make([]entity.MatchedUsers, 0)
	for _, user := range users.GetUsers() {
		finalUsers = append(finalUsers, entity.MatchedUsers{
			UserId:   user.UserId,
			Category: entity.MapToCategory(user.Category),
		})
	}
	return finalUsers
}
