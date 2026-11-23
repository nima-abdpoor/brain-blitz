package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
)

type RoundRecap struct {
	QuestionID    string `json:"questionId"`
	Question      string `json:"question"`
	PlayerAnswer  string `json:"playerAnswer"`
	CorrectAnswer string `json:"correctAnswer"`
	Answered      bool   `json:"answered"`
	IsCorrect     bool   `json:"isCorrect"`
}

func (svc Service) roundRecaps(ctx context.Context, gameID string, q GameQuestions) (map[uint64]RoundRecap, error) {
	board, err := svc.repository.GetLeaderBoard(ctx, gameID)
	if err != nil {
		return nil, err
	}
	question := q.Questions[q.CurrentQuestionIndex]
	recaps := map[uint64]RoundRecap{}
	for _, id := range q.Players {
		recaps[id] = recapForPlayer(question, playerAnswers(board, id))
	}
	return recaps, nil
}

func recapForPlayer(question Question, answers []QuestionCorrectness) RoundRecap {
	recap := RoundRecap{QuestionID: question.Id, Question: question.Content, CorrectAnswer: question.CorrectAnswer}
	for _, answer := range answers {
		if answer.QuestionId == question.Id {
			recap.PlayerAnswer = answer.PlayerChoice
			recap.Answered = true
			recap.IsCorrect = answer.PlayerChoice == question.CorrectAnswer
			break
		}
	}
	return recap
}
func (svc Service) sendRoundRecaps(gameID string, round int, recaps map[uint64]RoundRecap) {
	for id, recap := range recaps {
		svc.mu.RLock()
		conn := svc.connections[id]
		svc.mu.RUnlock()
		if conn == nil {
			continue
		}
		data, err := json.Marshal(ProcessGameMessageResponse{Success: true, Event: EventRoundClosed, Message: "round closed", MetaData: ProcessGameMetaDataResponse{GameId: gameID, RoundNumber: round, Recap: &recap}})
		if err == nil {
			err = svc.webSocket.WriteServerData(conn, websocket.OpText, string(data))
		}
		if err != nil {
			svc.logger.Warn("round recap delivery failed", "gameId", gameID, "playerId", id, "error", err.Error())
		}
	}
}
