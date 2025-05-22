package service

import "BrainBlitz.com/game/contract/match/golang"

func MapFromEntityToProtoMessage(matchedUsers []MatchedUsers) *golang.AllMatchedUsers {
	finalUsers := make([]*golang.MatchedUsers, 0)
	for _, user := range matchedUsers {
		categories := make([]string, 0)
		for _, category := range user.Category {
			categories = append(categories, string(category))
		}
		finalUsers = append(finalUsers, &golang.MatchedUsers{
			MatchId:  user.Id,
			UserId:   user.UserId,
			Category: categories,
		})
	}

	return &golang.AllMatchedUsers{
		Users: finalUsers,
	}
}

func MapFromAddToWaitingListProtoToEntity(message *golang.AddToWaitingList) (uint64, Category) {
	category := MapToCategory(message.GetCategory())
	userId := message.GetUserId()
	return userId, category
}
