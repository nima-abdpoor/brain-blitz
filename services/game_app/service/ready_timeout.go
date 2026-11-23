package service

import (
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"context"
	"errors"
	"time"
)

func (svc Service) scheduleReadyTimeout(ctx context.Context, gameID string, delay time.Duration) error {
	_, err := svc.taskPublisher.Publish(ctx, "game:ready-timeout", map[string]string{"gameId": gameID},
		taskqueue.MaxRetry(3), taskqueue.ProcessIn(delay))
	return err
}

func (svc Service) ProcessReadyTimeout(ctx context.Context, payload map[string]interface{}) error {
	gameID, ok := payload["gameId"].(string)
	if !ok || gameID == "" {
		return errors.New("ready timeout requires gameId")
	}
	lock := svc.roundMutex(gameID)
	lock.Lock()
	defer lock.Unlock()
	game, err := svc.repository.GetGame(ctx, gameID)
	if err != nil {
		return err
	}
	if game.Status != GameStatusPending || game.ReadyUntil.IsZero() {
		return nil
	}
	questions, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
	if err != nil && !errors.Is(err, ErrQuestionsUnavailable) {
		return err
	}
	if questions.RoundBased {
		return nil
	}
	if remaining := time.Until(game.ReadyUntil); remaining > 0 {
		// Asynq timestamps are stored at second precision.
		return svc.scheduleReadyTimeout(ctx, gameID, remaining+time.Second)
	}
	if err := svc.repository.CancelGame(ctx, gameID); err != nil {
		return err
	}
	if err := svc.writeMessage(game.Players, ProcessGameMessageResponse{
		Success: true, Event: EventMatchCancelled, Message: "ready deadline expired",
		MetaData: ProcessGameMetaDataResponse{GameId: gameID},
	}); err != nil {
		svc.logger.Debug("ready expiry notification incomplete", "gameId", gameID, "error", err)
	}
	return nil
}
