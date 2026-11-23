package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"BrainBlitz.com/game/pkg/matchqueue"
	"context"
	"net"
	"time"
)

type queueRepository interface {
	JoinQueue(context.Context, uint64, string, ...matchqueue.Choice) error
	LeaveQueue(context.Context, uint64) error
	IsQueued(context.Context, uint64, string, ...matchqueue.Choice) bool
}

func (svc Service) joinQueue(ctx context.Context, id uint64, category string, choices ...matchqueue.Choice) error {
	if r, ok := svc.repository.(queueRepository); ok {
		return r.JoinQueue(ctx, id, category, choices...)
	}
	return nil
}

func (svc Service) leaveQueue(id uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), saveUserGameStatusTimeOut)
	defer cancel()
	if err := svc.removeFromQueue(ctx, id); err != nil {
		svc.logger.Warn("queue cleanup failed", "playerId", id, "error", err)
	}
}

func (svc Service) removeFromQueue(ctx context.Context, id uint64) error {
	if r, ok := svc.repository.(queueRepository); ok {
		return r.LeaveQueue(ctx, id)
	}
	return nil
}

func (svc Service) acknowledgeQueueExit(ctx context.Context, id uint64, conn net.Conn) error {
	svc.matchmakingMu.Lock()
	defer svc.matchmakingMu.Unlock()
	if !svc.ownsConnection(id, conn) {
		return net.ErrClosed
	}
	err := svc.withCurrentMatch(ctx, id, func(match currentMatch) error {
		if err := svc.removeFromQueue(ctx, id); err != nil {
			return err
		}
		if match.active(time.Now()) {
			return svc.writeEvent(conn, websocket.OpText, matchCreatedEvent(match.id, match.game.ReadyUntil))
		}
		return svc.writeEvent(conn, websocket.OpText, ProcessGameMessageResponse{Success: true, Event: EventLeftQueue})
	})
	if err != nil {
		return svc.writeEvent(conn, websocket.OpText, ProcessGameMessageResponse{Event: EventLeftQueue})
	}
	return nil
}
