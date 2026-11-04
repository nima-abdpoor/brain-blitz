package service

import (
	"BrainBlitz.com/game/contract/event"
	"BrainBlitz.com/game/contract/match/golang"
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"time"
)

func (svc Service) ConsumeMatchCreated(message []byte, ctx context.Context) error {
	const op = "game.consumeMatchCreated"

	users := &golang.AllMatchedUsers{}
	err := proto.Unmarshal(message, users)
	if err != nil {
		svc.logger.Error(op, "error in unmarshalling match message", slog.String("error", err.Error()))
		return err
	}
	matchedUsers := MapFromProtoMessageToEntity(users)
	createdMatches := make([]MatchedUsers, 0)
	for _, matchedUser := range matchedUsers {
		if qr, ok := svc.repository.(queueRepository); ok {
			valid := true
			for _, id := range matchedUser.UserId {
				svc.mu.RLock()
				connected := svc.connections[id] != nil
				svc.mu.RUnlock()
				valid = valid && connected && len(matchedUser.Category) > 0 && qr.IsQueued(ctx, id, string(matchedUser.Category[0]))
			}
			if !valid {
				for _, id := range matchedUser.UserId {
					svc.mu.RLock()
					connected := svc.connections[id] != nil
					svc.mu.RUnlock()
					if connected && len(matchedUser.Category) > 0 && qr.IsQueued(ctx, id, string(matchedUser.Category[0])) {
						msg, _ := proto.Marshal(MapWaitingListRequestToProtoMessage(id, string(matchedUser.Category[0])))
						if err := svc.broker.Publish(ctx, event.GAME_V1_JOIN_MATCH_QUEUE_REQUESTED, msg); err != nil {
							return err
						}
					}
				}
				svc.logger.Info("discarded stale match", "matchId", matchedUser.MatchId)
				continue
			}
		}

		result, err := svc.repository.CreateGame(ctx, Game{
			Id:        nil,
			Players:   matchedUser.UserId,
			MatchId:   matchedUser.MatchId,
			Category:  matchedUser.Category,
			Status:    GameStatusPending,
			Question:  nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			svc.logger.Error(op, "error in creating match", slog.String("error", err.Error()))
		} else {
			matchedUser.GameId = result
			createdMatches = append(createdMatches, matchedUser)
		}

		svc.logger.Info(op, "game created", result)
	}

	for _, createdMatch := range createdMatches {
		go svc.saveUsersGameStatus(createdMatch.UserId, GameStatusPending)

		numberOfPlayers := len(createdMatch.UserId)
		go svc.saveGameStatus(createdMatch.GameId, nil, &numberOfPlayers)

		msg := ProcessGameMessageResponse{
			Success: true,
			Event:   EventMatchCreated,
			Message: "match created",
			MetaData: ProcessGameMetaDataResponse{
				GameId: createdMatch.GameId,
			},
		}
		err = svc.writeMessage(createdMatch.UserId, msg)
		if err != nil {
			svc.logger.Error(op, "error writing message", "userId", fmt.Sprintf("%v", createdMatch.UserId), "message", EventMatchCreated, slog.String("error", err.Error()))
		}
	}

	return nil
}
