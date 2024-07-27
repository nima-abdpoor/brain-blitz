package match

import entity "BrainBlitz.com/game/entity/game"

func MapFromEntityToProtoMessage(matchedUsers []entity.MatchedUsers) *AllMatchedUsers {
	finalUsers := make([]*MatchedUsers, len(matchedUsers))
	for _, user := range matchedUsers {
		finalUsers = append(finalUsers, &MatchedUsers{
			UserId:   user.UserId,
			Category: entity.MapFromCategory(user.Category),
		})
	}

	return &AllMatchedUsers{
		Users: finalUsers,
	}
}

func MapToEntityToProtoMessage(users *AllMatchedUsers) []entity.MatchedUsers {
	finalUsers := make([]entity.MatchedUsers, len(users.Users))
	for _, user := range users.GetUsers() {
		finalUsers = append(finalUsers, entity.MatchedUsers{
			UserId:   user.UserId,
			Category: entity.MapToCategory(user.Category),
		})
	}
	return finalUsers
}
