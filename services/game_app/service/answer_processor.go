package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"
)

func (svc Service) sendQuestionToPlayer(ctx context.Context, gameId string) error {
	op := "game.sendQuestionToPlayer"
	svc.logger.Info(op, "start sending questions to players", gameId)

	gameQuestions, err := svc.repository.GetQuestionsByGameId(context.Background(), gameId)
	if err != nil {
		return err
	}

	var questionsResponse []ProcessGameQuestion

	for _, questions := range gameQuestions.Questions {
		questionsResponse = append(questionsResponse, ProcessGameQuestion{
			Id:         questions.Id,
			Content:    questions.Content,
			Choices:    questions.Choices,
			Difficulty: questions.Difficulty,
			TTL:        questions.ValidAnswerTime,
		})
	}

	gameMessageResponse := ProcessGameMessageResponse{
		Success: true,
		Event:   NewQuestion,
		Message: "new question",
		MetaData: ProcessGameMetaDataResponse{
			GameId:    gameId,
			Questions: questionsResponse,
		},
	}

	jsonQuestions, err := json.Marshal(gameMessageResponse)
	if err != nil {
		return err
	}
	for _, playerId := range gameQuestions.Players {
		svc.mu.RLock()
		conn := svc.connections[playerId]
		svc.mu.RUnlock()
		err = svc.webSocket.WriteServerData(conn, websocket.OpText, string(jsonQuestions))
		if err != nil {
			svc.logger.Error(
				op, "error in sending questions to player",
				"playerId", playerId,
				"gameId", gameId,
				slog.String("error", err.Error()),
			)
		}
	}
	return nil
}

func (svc Service) savePlayerAnswer(ctx context.Context, playerId uint64, answer ProcessGameAnswer) (ProcessGameLeaderBoard, error) {
	op := "game.savePlayerAnswer"

	var leaderBoardResult ProcessGameLeaderBoard

	ps := PlayerAnswer{
		GameId:       answer.GameId,
		QuestionIDs:  answer.QuestionId,
		PlayerID:     strconv.FormatUint(playerId, 10),
		PlayerChoice: answer.Answer,
		AnswerTime:   time.Now().UTC(),
	}

	err := svc.repository.SavePlayerAnswer(ctx, answer.GameId, ps)
	if err != nil {
		svc.logger.Error(op, "error in saving answer of a player", slog.String("error", err.Error()))
		return leaderBoardResult, err
	}

	leaderBoard, err := svc.repository.GetLeaderBoard(ctx, answer.GameId)
	if err != nil {
		svc.logger.Error(op, "error in getting leader board", slog.String("error", err.Error()))
		return leaderBoardResult, err
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
	leaderBoardResult = ProcessGameLeaderBoard{
		GameId:      answer.GameId,
		PlayerPoint: playerPoints,
	}

	return leaderBoardResult, nil
}
