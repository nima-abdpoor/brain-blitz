package service

import "context"

type queueRepository interface {
	JoinQueue(context.Context, uint64, string) error
	LeaveQueue(context.Context, uint64) error
	IsQueued(context.Context, uint64, string) bool
}

func (svc Service) joinQueue(ctx context.Context, id uint64, category string) error {
	if r, ok := svc.repository.(queueRepository); ok {
		return r.JoinQueue(ctx, id, category)
	}
	return nil
}

func (svc Service) leaveQueue(id uint64) {
	if r, ok := svc.repository.(queueRepository); ok {
		ctx, cancel := context.WithTimeout(context.Background(), saveUserGameStatusTimeOut)
		defer cancel()
		if err := r.LeaveQueue(ctx, id); err != nil {
			svc.logger.Warn("queue cleanup failed", "playerId", id, "error", err)
		}
	}
}
