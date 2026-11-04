package service

import (
	"BrainBlitz.com/game/pkg/playerstats"
	"context"
)

func (svc Service) PlayerStatistics(ctx context.Context, userID uint64) (playerstats.Statistics, error) {
	return svc.repository.PlayerStatistics(ctx, userID)
}
