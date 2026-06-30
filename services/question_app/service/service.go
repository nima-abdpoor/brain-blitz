package service

import (
	"BrainBlitz.com/game/adapter/broker"
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/contract/match/golang"
	"BrainBlitz.com/game/pkg/logger"
	"context"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"time"
)

type Repository interface {
	GetProperQuestions(ctx context.Context, userId []uint64, category []Category, limit int) ([]Question, error)
	GetRandomQuestions(ctx context.Context, category []Category, difficulty Difficulty, limit int) ([]Question, error)
}

type Service struct {
	repository Repository
	broker     broker.Broker
	Logger     logger.Logger
}

func NewService(repository Repository, broker broker.Broker, logger logger.Logger) Service {
	return Service{
		repository: repository,
		broker:     broker,
		Logger:     logger,
	}
}

func (svc Service) AddQuestion(ctx context.Context, request AddQuestionRequest) (AddQuestionResponse, error) {
	return AddQuestionResponse{}, nil
}

func (svc Service) ConsumeMatchCreated(message []byte, ctx context.Context) error {
	questionsTopic := event.QUESTION_V1_QUESTIONS
	const op = "service.consumeMatchCreated"
	const limit = 10

	protoMatchedUsers := &golang.AllMatchedUsers{}
	err := proto.Unmarshal(message, protoMatchedUsers)
	if err != nil {
		svc.Logger.Error(op, "error in unmarshalling match message", slog.String("error", err.Error()))
		return err
	}
	matchedUsers := MapFromProtoMessageToEntity(protoMatchedUsers)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, matchedUser := range matchedUsers {
		svc.Logger.Info(op, "userId", matchedUser.UserId, "category", matchedUser.Category, "limit", limit)
		questions, err := svc.repository.GetRandomQuestions(ctx, matchedUser.Category, DifficultEasy, limit)
		if err != nil {
			svc.Logger.Error(op, "error in getting matched questions", err)
		}

		buff, err := proto.Marshal(MapQuestionsToProtoMessage(matchedUser.MatchId, questions))
		if err != nil {
			svc.Logger.Error(op, "message", "error in marshaling questions message", err.Error())
		}

		err = svc.broker.Publish(ctx, questionsTopic, buff)
		if err != nil {
			svc.Logger.Error(op, "message", "error in publishing questions message", err.Error())
		}
		svc.Logger.Info(op, "message", "publishing questions...", "questions", questions, "matchId", matchedUser.MatchId)
	}

	return nil
}
