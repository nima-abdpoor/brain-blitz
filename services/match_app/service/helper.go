package service

import "BrainBlitz.com/game/contract/match/golang"

func MapFromEntityToProtoMessage(matchedUsers []MatchedUsers) *golang.AllMatchedUsers {
	finalUsers := make([]*golang.MatchedUsers, 0)
	for _, user := range matchedUsers {
		categories := make([]string, 0)
		for _, category := range user.Category {
			categories = append(categories, MapFromCategory(category))
		}
		finalUsers = append(finalUsers, &golang.MatchedUsers{
			UserId:   user.UserId,
			Category: categories,
		})
	}

	return &golang.AllMatchedUsers{
		Users: finalUsers,
	}
}
