package service

import (
	"context"
	"log/slog"
)

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
