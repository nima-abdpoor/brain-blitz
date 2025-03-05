package golang

import (
	"BrainBlitz.com/game/match_app/service"
)

func MapFromEntityToProtoMessage(matchedUsers []service.MatchedUsers) *AllMatchedUsers {
	finalUsers := make([]*MatchedUsers, 0)
	for _, user := range matchedUsers {
		finalUsers = append(finalUsers, &MatchedUsers{
			UserId:   user.UserId,
			Category: service.MapFromCategory(user.Category),
		})
	}

	return &AllMatchedUsers{
		Users: finalUsers,
	}
}

func MapToEntityToProtoMessage(users *AllMatchedUsers) []service.MatchedUsers {
	finalUsers := make([]service.MatchedUsers, 0)
	for _, user := range users.GetUsers() {
		finalUsers = append(finalUsers, service.MatchedUsers{
			UserId:   user.UserId,
			Category: service.MapToCategory(user.Category),
		})
	}
	return finalUsers
}
