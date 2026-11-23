package service

import (
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/contract/match/golang"
	"BrainBlitz.com/game/pkg/matchqueue"
	"context"
	"errors"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"
)

func (svc Service) ConsumeMatchCreated(message []byte, ctx context.Context) error {
	users := &golang.AllMatchedUsers{}
	if err := proto.Unmarshal(message, users); err != nil {
		return err
	}
	var createErr error
	for _, match := range MapFromProtoMessageToEntity(users) {
		createErr = errors.Join(createErr, svc.createMatch(ctx, match))
	}
	return createErr
}

func (svc Service) createMatch(ctx context.Context, match MatchedUsers) error {
	svc.matchmakingMu.Lock()
	defer svc.matchmakingMu.Unlock()
	if qr, ok := svc.repository.(queueRepository); ok {
		if len(match.Category) == 0 {
			return nil
		}
		category := string(match.Category[0])
		choice := matchqueue.Choice{Seconds: match.QuestionSeconds, OptionID: match.TimeOptionID, Players: len(match.UserId)}
		var queued []uint64
		for _, id := range match.UserId {
			svc.mu.RLock()
			connected := svc.connections[id] != nil
			svc.mu.RUnlock()
			if connected && qr.IsQueued(ctx, id, category, choice) {
				queued = append(queued, id)
			}
		}
		if len(queued) != len(match.UserId) {
			for _, id := range queued {
				msg, err := proto.Marshal(MapWaitingListRequestToProtoMessage(id, category, choice))
				if err != nil {
					return err
				}
				if err := svc.broker.Publish(ctx, event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED, msg); err != nil {
					return err
				}
			}
			svc.logger.Info("discarded stale match", "matchId", match.MatchId)
			return nil
		}
	}

	names := make(map[string]string)
	svc.mu.RLock()
	for _, id := range match.UserId {
		names[strconv.FormatUint(id, 10)] = svc.playerNames[id]
	}
	svc.mu.RUnlock()
	now := time.Now().UTC()
	deadline := now.Add(svc.config.ReadyTimeout)
	gameID, err := svc.repository.CreateGame(ctx, Game{
		QuestionSeconds: match.QuestionSeconds, TimeOptionID: match.TimeOptionID,
		Players: match.UserId, PlayerNames: names, MatchId: match.MatchId, Category: match.Category,
		Status: GameStatusPending, CreatedAt: now, UpdatedAt: now, ReadyUntil: deadline,
	})
	if err != nil {
		return svc.rejectMatchedQueue(match.UserId, err)
	}
	lock := svc.roundMutex(gameID)
	lock.Lock()
	defer lock.Unlock()
	count := len(match.UserId)
	_, err = svc.repository.UpsertReadyPlayer(ctx, gameID, nil, &count)
	if err == nil {
		err = svc.scheduleReadyTimeout(ctx, gameID, time.Until(deadline))
	}
	if err != nil {
		// No client has been told about the match. Release the roster to requeue.
		svc.logger.Error("could not initialize match readiness", "gameId", gameID, "error", err)
		return svc.rejectMatchedQueue(match.UserId, errors.Join(err, svc.repository.CancelGame(ctx, gameID)))
	}

	for _, id := range match.UserId {
		svc.leaveQueue(id)
	}
	svc.saveUsersGameStatus(match.UserId, GameStatusPending)
	if err := svc.writeMessage(match.UserId, matchCreatedEvent(gameID, deadline)); err != nil {
		svc.logger.Warn("match notification incomplete", "gameId", gameID, "error", err)
	}
	return nil
}

func (svc Service) rejectMatchedQueue(players []uint64, err error) error {
	for _, id := range players {
		svc.leaveQueue(id)
	}
	return errors.Join(err, svc.writeMessage(players, ProcessGameMessageResponse{Event: Error, Message: "could not create match"}))
}

func matchCreatedEvent(gameID string, deadline time.Time) ProcessGameMessageResponse {
	return ProcessGameMessageResponse{
		Success: true, Event: EventMatchCreated, Message: "match created",
		MetaData: ProcessGameMetaDataResponse{GameId: gameID, ReadyUntil: &deadline},
	}
}
