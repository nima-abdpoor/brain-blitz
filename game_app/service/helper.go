package service

import "BrainBlitz.com/game/contract/match/golang"

func MapFromProtoMessageToEntity(users *golang.AllMatchedUsers) []MatchedUsers {
	finalUsers := make([]MatchedUsers, 0)
	for _, user := range users.GetUsers() {
		categories := make([]Category, 0)
		for _, category := range user.GetCategory() {
			categories = append(categories, MapToCategory(category))
		}
		finalUsers = append(finalUsers, MatchedUsers{
			UserId:   user.UserId,
			Category: categories,
		})
	}
	return finalUsers
}
