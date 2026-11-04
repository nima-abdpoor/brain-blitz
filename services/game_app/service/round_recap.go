package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
	"strconv"
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
		recap := RoundRecap{QuestionID: question.Id, Question: question.Content, CorrectAnswer: question.CorrectAnswer}
		for _, p := range board.PlayersPoint {
			if p.PlayerId == strconv.FormatUint(id, 10) {
				for _, a := range p.QuestionCorrectness {
					if a.QuestionId == question.Id {
						recap.PlayerAnswer = a.PlayerChoice
						recap.Answered = true
						recap.IsCorrect = a.PlayerChoice == question.CorrectAnswer
					}
				}
			}
		}
		recaps[id] = recap
	}
	return recaps, nil
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
