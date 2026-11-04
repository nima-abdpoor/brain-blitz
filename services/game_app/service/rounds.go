package service

import (
	taskqueue "BrainBlitz.com/game/adapter/task-queue"
	"context"
	"fmt"
	"strconv"
	"time"
)

// Round updates are serialized with answer acceptance. Timeout tasks include
// their round index, so a task for a round closed early is a harmless no-op.
func (svc Service) startRounds(ctx context.Context, gameID string) error {
	mu := svc.roundMutex(gameID)
	mu.Lock()
	defer mu.Unlock()
	q, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
	if err == nil && q.RoundBased {
		return nil
	}
	if _, err := svc.prepareGameQuestions(ctx, gameID); err != nil {
		return err
	}
	return svc.openRound(ctx, gameID, 0)
}

func (svc Service) openRound(ctx context.Context, gameID string, index int) error {
	var recaps map[uint64]RoundRecap
	if index > 0 {
		previous, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
		if err != nil {
			return err
		}
		if !previous.Completed && previous.CurrentQuestionIndex == index-1 {
			recaps, err = svc.roundRecaps(ctx, gameID, previous)
			if err != nil {
				return err
			}
		}
	}
	q, err := svc.repository.OpenRound(ctx, gameID, index)
	if err != nil {
		return err
	}
	svc.sendRoundRecaps(gameID, index, recaps)
	if q.Completed {
		return svc.ProcessGameCompletion(ctx, map[string]interface{}{"gameId": gameID})
	}
	// Schedule before publishing: a delivery failure must not leave an
	// unanswered round without a deadline. Offline players time out normally.
	_, err = svc.taskPublisher.Publish(ctx, "game:round-timeout", struct {
		GameID string `json:"gameId"`
		Round  int    `json:"round"`
	}{gameID, index}, taskqueue.MaxRetry(3), taskqueue.ProcessIn(time.Until(q.Questions[index].ValidAnswerTime)))
	if err != nil {
		return err
	}
	svc.logger.Info("round opened", "gameId", gameID, "round", index+1, "deadline", q.Questions[index].ValidAnswerTime)
	return svc.sendQuestionToPlayer(ctx, gameID)
}

func (svc Service) ProcessRoundTimeout(ctx context.Context, payload map[string]interface{}) error {
	gameID, ok := payload["gameId"].(string)
	if !ok {
		return fmt.Errorf("round timeout requires gameId")
	}
	round, ok := payload["round"].(float64)
	if !ok {
		return fmt.Errorf("round timeout requires round")
	}
	// Redis task timestamps have second precision. An early dequeue must not
	// enter Asynq's long failure backoff or block answers behind roundMu.
	initial, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
	if err != nil {
		return err
	}
	if initial.Completed {
		mu := svc.roundMutex(gameID)
		mu.Lock()
		defer mu.Unlock()
		return svc.ProcessGameCompletion(ctx, payload)
	}
	if !initial.RoundBased || initial.CurrentQuestionIndex != int(round) || initial.CurrentQuestionIndex >= len(initial.Questions) {
		return nil
	}
	if !initial.Completed && initial.CurrentQuestionIndex == int(round) {
		remaining := time.Until(initial.Questions[initial.CurrentQuestionIndex].ValidAnswerTime)
		if remaining > time.Second {
			_, err := svc.taskPublisher.Publish(ctx, "game:round-timeout", payload, taskqueue.MaxRetry(3), taskqueue.ProcessIn(remaining+time.Second))
			return err
		}
		if remaining > 0 {
			timer := time.NewTimer(remaining)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	mu := svc.roundMutex(gameID)
	mu.Lock()
	defer mu.Unlock()
	q, err := svc.repository.GetQuestionsByGameId(ctx, gameID)
	if err != nil {
		return err
	}
	if q.Completed || !q.RoundBased || q.CurrentQuestionIndex != int(round) {
		return nil
	}
	if time.Now().Before(q.Questions[q.CurrentQuestionIndex].ValidAnswerTime) {
		return fmt.Errorf("round deadline has not elapsed")
	}
	return svc.openRound(ctx, gameID, q.CurrentQuestionIndex+1)
}

func (svc Service) acceptRoundAnswer(ctx context.Context, playerID uint64, answer ProcessGameAnswer) (ProcessGameLeaderBoard, bool, error) {
	q, err := svc.repository.GetQuestionsByGameId(ctx, answer.GameId)
	if err != nil {
		return ProcessGameLeaderBoard{}, false, err
	}
	if q.RoundBased {
		if q.Completed || q.CurrentQuestionIndex >= len(q.Questions) || q.Questions[q.CurrentQuestionIndex].Id != answer.QuestionId {
			return ProcessGameLeaderBoard{}, false, fmt.Errorf("question is no longer active")
		}
		member := false
		for _, id := range q.Players {
			member = member || id == playerID
		}
		if !member {
			return ProcessGameLeaderBoard{}, false, fmt.Errorf("player is not in this game")
		}
	}
	board, err := svc.savePlayerAnswer(ctx, playerID, answer)
	if err != nil {
		return board, false, err
	}
	answered := map[string]bool{}
	for _, player := range board.PlayerPoint {
		for _, a := range player.Answers {
			if a.QuestionId == answer.QuestionId {
				answered[player.PlayerId] = true
			}
		}
	}
	all := q.RoundBased && len(q.Players) > 0
	for _, id := range q.Players {
		all = all && answered[strconv.FormatUint(id, 10)]
	}
	svc.logger.Info("answer accepted", "gameId", answer.GameId, "questionId", answer.QuestionId, "allAnswered", all)
	return board, all, nil
}

func (svc Service) addPlayerNames(board ProcessGameLeaderBoard) ProcessGameLeaderBoard {
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	for i := range board.PlayerPoint {
		id, _ := strconv.ParseUint(board.PlayerPoint[i].PlayerId, 10, 64)
		board.PlayerPoint[i].DisplayName = svc.playerNames[id]
	}
	return board
}
