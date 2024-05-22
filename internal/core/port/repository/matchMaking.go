package repository

import (
	entity "BrainBlitz.com/game/entity/game"
	"context"
)

type MatchMakingRepository interface {
	AddToWaitingList(category entity.Category, userId string) error
	GetWaitingListByCategory(ctx context.Context, category entity.Category) ([]entity.WaitingMember, error)
}

type PresenceClient interface {
	GetPresenceByUserID(context context.Context, userId string) (int64, error)
}
