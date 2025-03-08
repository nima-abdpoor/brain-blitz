package service

import (
	entity "BrainBlitz.com/game/entity/game"
	"context"
)

type Repository interface {
	CreateMatch(ctx context.Context, game entity.Game) (string, error)
}
