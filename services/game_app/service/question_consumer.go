package service

import (
	questionProto "BrainBlitz.com/game/contract/question/golang"
	"context"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

func (svc Service) ConsumeQuestions(message []byte, ctx context.Context) error {
	const op = "game.consumeQuestions"

	protoQuestions := &questionProto.Questions{}
	err := proto.Unmarshal(message, protoQuestions)
	if err != nil {
		svc.logger.Error(op, "error in unmarshalling question message", slog.String("error", err.Error()))
		return err
	}
	questions, matchId := MapFromProtoMessageToQuestionsEntity(protoQuestions)

	svc.logger.Info(op, "question list received", questions, "matchId", matchId)

	err = svc.repository.SaveQuestionsByMatchId(ctx, matchId, questions)
	if err != nil {
		svc.logger.Error(op, "error in saving Questions", slog.String("error", err.Error()))
	}

	return nil
}
