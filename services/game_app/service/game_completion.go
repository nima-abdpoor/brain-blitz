package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
	"fmt"
)

func (svc Service) ProcessGameCompletion(ctx context.Context, payload map[string]interface{}) error {
	gameID, ok := payload["gameId"].(string)
	if !ok || gameID == "" {
		return fmt.Errorf("game completion payload requires gameId")
	}
	game, err := svc.repository.GetGame(ctx, gameID)
	if err != nil {
		return err
	}
	leaderboard, err := svc.repository.GetLeaderBoard(ctx, gameID)
	if err != nil {
		return err
	}

	points := make([]ProcessGamePlayerPoint, 0, len(leaderboard.PlayersPoint))
	for _, player := range leaderboard.PlayersPoint {
		answers := make([]ProcessGameAnswerResult, 0, len(player.QuestionCorrectness))
		for _, answer := range player.QuestionCorrectness {
			answers = append(answers, ProcessGameAnswerResult{
				QuestionId: answer.QuestionId, CorrectAnswer: answer.CorrectChoice,
				PlayerAnswer: answer.PlayerChoice, IsCorrect: answer.IsCorrect,
			})
		}
		points = append(points, ProcessGamePlayerPoint{PlayerId: player.PlayerId, Point: player.Point, Answers: answers})
	}
	response, err := json.Marshal(ProcessGameMessageResponse{
		Success: true, Event: EventCompleted, Message: "game completed",
		MetaData: ProcessGameMetaDataResponse{
			GameId: gameID,
			Answer: ProcessGameLeaderBoard{GameId: gameID, PlayerPoint: points},
		},
	})
	if err != nil {
		return err
	}
	if err := svc.repository.FinishGame(ctx, gameID); err != nil {
		return err
	}

	// Notify the full roster, including players who did not submit an answer.
	// Disconnected players still get their state reset so they can play again.
	for _, playerID := range game.Players {
		if err := svc.repository.UpsertUserStatus(ctx, playerID, GameStatusFinished); err != nil {
			return err
		}
		svc.mu.RLock()
		conn := svc.connections[playerID]
		svc.mu.RUnlock()
		if conn == nil {
			continue
		}
		if err := svc.webSocket.WriteServerData(conn, websocket.OpText, string(response)); err != nil {
			svc.logger.Warn("game completion delivery failed", "playerId", playerID, "error", err.Error())
		}
		svc.closeConnection(playerID, conn)
	}
	return nil
}
