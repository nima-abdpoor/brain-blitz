package service

import (
	matchProto "BrainBlitz.com/game/contract/match/golang"
	questionProto "BrainBlitz.com/game/contract/question/golang"
	"BrainBlitz.com/game/pkg/matchqueue"
)

func MapFromProtoMessageToEntity(users *matchProto.AllMatchedUsers) []MatchedUsers {
	finalUsers := make([]MatchedUsers, 0)
	for _, user := range users.GetUsers() {
		categories := make([]Category, 0)
		for _, category := range user.GetCategory() {
			categories = append(categories, MapToCategory(category))
		}
		finalUsers = append(finalUsers, MatchedUsers{
			MatchId:         user.GetMatchId(),
			QuestionSeconds: int(user.GetQuestionSeconds()), TimeOptionID: user.GetTimeOptionId(),
			UserId:   user.GetUserId(),
			Category: categories,
		})
	}
	return finalUsers
}

func MapFromProtoMessageToQuestionsEntity(questions *questionProto.Questions) ([]Question, string) {
	finalQuestions := make([]Question, 0)

	for _, question := range questions.GetQuestions() {
		finalQuestions = append(finalQuestions, Question{
			Id:            question.GetQuestionId(),
			Content:       question.GetContent(),
			CorrectAnswer: question.GetCorrectAnswer(),
			Choices:       question.GetChoices(),
			Category:      MapToCategory(question.GetCategory()),
			Difficulty:    MapToDifficulty(question.GetDifficulty()),
		})
	}

	return finalQuestions, questions.GetMatchId()
}

func MapWaitingListRequestToProtoMessage(userId uint64, category string, choices ...matchqueue.Choice) *matchProto.AddToWaitingList {
	var choice matchqueue.Choice
	if len(choices) > 0 {
		choice = choices[0]
	}
	return &matchProto.AddToWaitingList{
		QuestionSeconds: int32(choice.Seconds), TimeOptionId: choice.OptionID,
		UserId:   userId,
		Category: category,
	}
}
