package repository

import (
	entity "BrainBlitz.com/game/entity/game"
	"context"
)

type MatchManagementRepository interface {
	CreateMatch(ctx context.Context, game entity.Game) error
}
