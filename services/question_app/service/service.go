package service

import (
	"BrainBlitz.com/game/pkg/logger"
	"context"
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
