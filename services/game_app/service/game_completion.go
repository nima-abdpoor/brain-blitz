package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
)

func (svc Service) ProcessGameCompletion(ctx context.Context, payload map[string]interface{}) error {
	op := "service.ProcessGameCompletion"

	gameId := payload["gameId"].(string)

	leaderBoard, err := svc.repository.GetLeaderBoard(ctx, gameId)
	if err != nil {
		svc.logger.Error(op, "error in getting leader board", slog.String("error", err.Error()))
		gameInfo, err := svc.repository.GetGame(ctx, gameId)
		if err != nil {
			svc.logger.Error(op, "error in getting game info", "gameId", gameId, slog.String("error", err.Error()))
			return err
		}

		for _, playerId := range gameInfo.Players {
			svc.mu.RLock()
			conn := svc.connections[playerId]
			svc.mu.RUnlock()
			response := ProcessGameMessageResponse{
				Success: false,
				Event:   EventCompleted,
				Message: "internal server error",
			}
			jsonResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}
			err = svc.webSocket.WriteServerData(conn, websocket.OpText, string(jsonResponse))
			if err != nil {
				return err
			}
		}
		return err
	}

	var playerPoints []ProcessGamePlayerPoint

	for _, playerPoint := range leaderBoard.PlayersPoint {
		var ansResult []ProcessGameAnswerResult
		for _, questionAnswers := range playerPoint.QuestionCorrectness {
			ansResult = append(ansResult, ProcessGameAnswerResult{
				QuestionId:    questionAnswers.QuestionId,
				CorrectAnswer: questionAnswers.CorrectChoice,
				PlayerAnswer:  questionAnswers.PlayerChoice,
				IsCorrect:     questionAnswers.IsCorrect,
			})
		}
		playerPoints = append(playerPoints, ProcessGamePlayerPoint{
			PlayerId: playerPoint.PlayerId,
			Point:    playerPoint.Point,
			Answers:  ansResult,
		})
	}
	leaderBoardResult := ProcessGameLeaderBoard{
		GameId:      gameId,
		PlayerPoint: playerPoints,
	}

	msg := ProcessGameMessageResponse{
		Success: true,
		Event:   EventCompleted,
		Message: "game completed",
		MetaData: ProcessGameMetaDataResponse{
			GameId: gameId,
			Answer: leaderBoardResult,
		},
	}

	jsonResponse, err := json.Marshal(msg)
	if err != nil {
		svc.logger.Info(op, "error in converting struct to json", "error", err.Error())
	}

	for _, playerPoint := range playerPoints {
		playerId, err := strconv.ParseInt(playerPoint.PlayerId, 10, 64)
		if err != nil {
			svc.logger.Info(op, "error in converting playerId from string to int", "playerId", playerPoint.PlayerId, "error", err.Error())
			return err
		}
		svc.mu.RLock()
		conn := svc.connections[uint64(playerId)]
		svc.mu.RUnlock()
		err = svc.webSocket.WriteServerData(conn, websocket.OpText, string(jsonResponse))
		if err != nil {
			svc.logger.Error(op, "error writing message", "playerId", fmt.Sprintf("%v", playerId), "message", EventCompleted, slog.String("error", err.Error()))
			return err
		}
		svc.mu.Lock()
		delete(svc.connections, uint64(playerId))
		svc.mu.Unlock()
		conn.Close()
	}

	return nil
}
