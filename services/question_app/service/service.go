package service

import (
	"BrainBlitz.com/game/contract/match/golang"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

type Repository interface {
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

	users := &golang.AllMatchedUsers{}
	err := proto.Unmarshal(message, users)
	if err != nil {
		svc.Logger.Error(op, "error in unmarshalling match message", slog.String("error", err.Error()))
		return err
	}
	matchedUsers := MapFromProtoMessageToEntity(users)
	for _, matchedUser := range matchedUsers {
		svc.Logger.Info(op, "matched user", matchedUser)
	}

	return nil
}
