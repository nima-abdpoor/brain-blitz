package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Matching and question delivery arrive on separate Kafka consumers. A client
// may send READY before the question consumer has populated the game cache.
func (svc Service) prepareGameQuestions(ctx context.Context, gameID string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return 0, fmt.Errorf("waiting for game questions: %w", err)
		}
		questions, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
		if err == nil && len(questions.Questions) > 0 {
			if questions.Completed {
				return 0, fmt.Errorf("game already completed")
			}
			if questions.RoundBased {
				return time.Until(questions.Questions[questions.CurrentQuestionIndex].ValidAnswerTime), nil
			}
			return svc.repository.SetValidAnswerTimeForQuestions(ctx, gameID)
		}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("waiting for game questions: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (svc Service) saveUsersGameStatus(userId []uint64, status GameStatus) {
	const op = "game.saveUsersGameStatus"

	for _, id := range userId {
		go func(id uint64) {
			upsertUserStatusCtx, cancel := context.WithTimeout(context.Background(), saveUserGameStatusTimeOut)
			defer cancel()

			if err := svc.repository.UpsertUserStatus(upsertUserStatusCtx, id, status); err != nil {
				svc.logger.Error(op, "error in saving user status", "error", err.Error())
			}
		}(id)
	}
}

func (svc Service) saveGameStatus(gameId string, userId *uint64, numberOfPlayers *int) bool {
	const op = "game.saveGameStatus"

	ctx, cancel := context.WithTimeout(context.Background(), saveGameStatusTimeOut)
	defer cancel()

	var id *int
	if userId != nil {
		uId := int(*userId)
		id = &uId
	}
	isGameReady, err := svc.repository.UpsertReadyPlayer(ctx, gameId, id, numberOfPlayers)
	if err != nil {
		svc.logger.Error(op, "error in saving ready player")
	}

	return isGameReady
}

func (svc Service) markReady(ctx context.Context, gameID string, playerID uint64) (bool, error) {
	game, err := svc.repository.GetGame(ctx, gameID)
	if err != nil {
		return false, err
	}
	for _, member := range game.Players {
		if member == playerID {
			id := int(playerID)
			return svc.repository.UpsertReadyPlayer(ctx, gameID, &id, nil)
		}
	}
	return false, fmt.Errorf("player is not in this game")
}

func (svc Service) getUsersGameStatus(userId uint64) GameStatus {
	const op = "game.getUsersGameStatus"
	ctx, cancel := context.WithTimeout(context.Background(), svc.config.StoreGameStatusTimeOut)
	defer cancel()
	status, err := svc.repository.GetUserStatus(ctx, userId)
	if err != nil {
		svc.logger.Error(op, "error in getting user status", slog.String("error", err.Error()))
		return GameStatusUnknown
	}
	return status
}
