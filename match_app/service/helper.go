package service

import "BrainBlitz.com/game/contract/match/golang"

func MapFromEntityToProtoMessage(matchedUsers []MatchedUsers) *golang.AllMatchedUsers {
	finalUsers := make([]*golang.MatchedUsers, 0)
	for _, user := range matchedUsers {
		finalUsers = append(finalUsers, &golang.MatchedUsers{
			UserId:   user.UserId,
			Category: MapFromCategory(user.Category),
		})
	}

	return &golang.AllMatchedUsers{
		Users: finalUsers,
	}
}

func MapToEntityToProtoMessage(users *golang.AllMatchedUsers) []MatchedUsers {
	finalUsers := make([]MatchedUsers, 0)
	for _, user := range users.GetUsers() {
		finalUsers = append(finalUsers, MatchedUsers{
			UserId:   user.UserId,
			Category: MapToCategory(user.Category),
		})
	}
	return finalUsers
}
