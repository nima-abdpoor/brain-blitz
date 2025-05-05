package service

import (
	"BrainBlitz.com/game/contract/match/golang"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

type Repository interface {
	GetProperQuestions(ctx context.Context, userId []uint64, category []Category, limit int) ([]Question, error)
	GetRandomQuestions(ctx context.Context, category []Category, difficulty Difficulty, limit int) ([]Question, error)
}

type Service struct {
	repository Repository
	Logger     logger.Logger
}

func NewService(repository Repository, logger logger.Logger) Service {
	return Service{
		repository: repository,
		Logger:     logger,
	}
}

func (svc Service) AddQuestion(ctx context.Context, request AddQuestionRequest) (AddQuestionResponse, error) {
	return AddQuestionResponse{}, nil
}

func (svc Service) ConsumeMatchCreated(message []byte, ctx context.Context) error {
	const op = "service.consumeMatchCreated"
	const limit = 10

	protoMatchedUsers := &golang.AllMatchedUsers{}
	err := proto.Unmarshal(message, protoMatchedUsers)
	if err != nil {
		svc.Logger.Error(op, "error in unmarshalling match message", slog.String("error", err.Error()))
		return err
	}
	matchedUsers := MapFromProtoMessageToEntity(protoMatchedUsers)
	for _, matchedUser := range matchedUsers {
		svc.Logger.Info(op, "userId", matchedUser.UserId, "category", matchedUser.Category, "limit", limit)
		questions, err := svc.repository.GetRandomQuestions(ctx, matchedUser.Category, DifficultEasy, limit)
		if err != nil {
			svc.Logger.Error(op, "error in getting matched questions", err)
		}

		// todo publish questions
		svc.Logger.Info(op, "matched user", matchedUser)
		svc.Logger.Info(op, "Questions", questions, "matchId", "id")
	}

	return nil
}
