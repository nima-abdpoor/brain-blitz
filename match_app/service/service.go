package service

import (
	"context"
)

type Repository interface {
	AddToWaitingList(ctx context.Context, category Category, userId string) error
	GetWaitingListByCategory(ctx context.Context, category Category) ([]WaitingMember, error)
}
