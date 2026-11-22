package service

import (
	"BrainBlitz.com/game/adapter/websocket"
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"time"
)

var ErrQuestionsUnavailable = errors.New("game questions have not been cached")

const resultRecoveryWindow = 24 * time.Hour

type MatchSnapshot struct {
	GameID          string                   `json:"gameId"`
	Category        Category                 `json:"category"`
	TimeOptionID    string                   `json:"timeOptionId"`
	QuestionSeconds int                      `json:"questionSeconds"`
	Players         int                      `json:"players"`
	Phase           string                   `json:"phase"`
	Ready           bool                     `json:"ready"`
	ReadyUntil      *time.Time               `json:"readyUntil,omitempty"`
	RoundNumber     int                      `json:"roundNumber,omitempty"`
	TotalQuestions  int                      `json:"totalQuestions,omitempty"`
	Question        *ProcessGameQuestion     `json:"question,omitempty"`
	Submission      *RecoveredSubmission     `json:"submission,omitempty"`
	Recap           *RoundRecap              `json:"recap,omitempty"`
	RecapRound      int                      `json:"recapRound,omitempty"`
	Scores          []ProcessGamePlayerPoint `json:"scores,omitempty"`
}

type RecoveredSubmission struct {
	Choice    string `json:"choice"`
	IsCorrect bool   `json:"isCorrect"`
}

type currentMatch struct {
	id        string
	game      Game
	questions GameQuestions
}

func (m currentMatch) active(now time.Time) bool {
	return m.id != "" && m.game.Status != GameStatusFinished && m.game.Status != GameStatusCancelled &&
		(m.questions.RoundBased || now.Before(m.game.ReadyUntil))
}

// Hold matchmakingMu across the lookup and the callback. This orders the
// no-match response with match creation. The round lock orders the response
// with READY, ANSWER, timeouts and all round or completion broadcasts.
func (svc Service) withCurrentMatch(ctx context.Context, playerID uint64, visit func(currentMatch) error) error {
	gameID, err := svc.repository.LatestGameID(ctx, playerID)
	if err != nil {
		return err
	}
	if gameID == "" {
		return visit(currentMatch{})
	}
	lock := svc.roundMutex(gameID)
	lock.Lock()
	defer lock.Unlock()
	game, err := svc.repository.GetGame(ctx, gameID)
	if err != nil {
		return err
	}
	match := currentMatch{id: gameID, game: game}
	if game.Status != GameStatusFinished && game.Status != GameStatusCancelled {
		match.questions, err = svc.repository.GetQuestionsByGameId(ctx, gameID)
		if err != nil && !errors.Is(err, ErrQuestionsUnavailable) {
			return err
		}
		if game.ReadyUntil.IsZero() {
			match.game.ReadyUntil = game.CreatedAt.Add(svc.config.ReadyTimeout)
		}
	}
	return visit(match)
}

func (svc Service) syncMatch(ctx context.Context, playerID uint64, conn net.Conn) error {
	svc.matchmakingMu.Lock()
	defer svc.matchmakingMu.Unlock()
	err := svc.withCurrentMatch(ctx, playerID, func(match currentMatch) error {
		if !svc.ownsConnection(playerID, conn) {
			return net.ErrClosed
		}
		snapshot, err := svc.matchSnapshot(ctx, match, playerID)
		if err != nil {
			return err
		}
		if snapshot == nil || snapshot.Phase == "finished" {
			if err := svc.removeFromQueue(ctx, playerID); err != nil {
				return err
			}
		}
		return svc.writeEvent(conn, websocket.OpText, ProcessGameMessageResponse{
			Success: true, Event: EventMatchSnapshot, MetaData: ProcessGameMetaDataResponse{Snapshot: snapshot},
		})
	})
	if err != nil && !errors.Is(err, net.ErrClosed) {
		svc.logger.Warn("could not load match snapshot", "playerId", playerID, "error", err)
		return svc.writeEvent(conn, websocket.OpText, ProcessGameMessageResponse{Event: EventMatchSnapshot})
	}
	return err
}

func (svc Service) matchSnapshot(ctx context.Context, match currentMatch, playerID uint64) (*MatchSnapshot, error) {
	game, q := match.game, match.questions
	now := time.Now()
	finished := game.Status == GameStatusFinished && now.Sub(game.UpdatedAt) <= resultRecoveryWindow
	if !finished && !match.active(now) {
		return nil, nil
	}
	result := &MatchSnapshot{
		GameID: match.id, TimeOptionID: game.TimeOptionID, QuestionSeconds: game.QuestionSeconds, Players: len(game.Players),
	}
	if len(game.Category) > 0 {
		result.Category = game.Category[0]
	}
	if !finished && !q.RoundBased {
		ready, err := svc.repository.IsPlayerReady(ctx, match.id, playerID)
		result.Phase, result.Ready, result.ReadyUntil = "ready", ready, &game.ReadyUntil
		return result, err
	}
	board, err := svc.repository.GetLeaderBoard(ctx, match.id)
	if err != nil {
		return nil, err
	}
	if finished || q.Completed {
		result.Phase = "finished"
		result.Scores = svc.leaderBoardResponse(match.id, game, board).PlayerPoint
		for i := range result.Scores {
			result.Scores[i].Answers = nil
		}
		return result, nil
	}
	if q.CurrentQuestionIndex < 0 || q.CurrentQuestionIndex >= len(q.Questions) {
		return nil, errors.New("invalid active question index")
	}
	current := q.Questions[q.CurrentQuestionIndex]
	result.Phase = "round"
	result.RoundNumber, result.TotalQuestions = q.CurrentQuestionIndex+1, len(q.Questions)
	question := publicQuestion(current)
	result.Question = &question
	answers := playerAnswers(board, playerID)
	for _, answer := range answers {
		if answer.QuestionId == current.Id {
			result.Submission = &RecoveredSubmission{Choice: answer.PlayerChoice, IsCorrect: answer.PlayerChoice == current.CorrectAnswer}
			break
		}
	}
	if q.CurrentQuestionIndex > 0 {
		recap := recapForPlayer(q.Questions[q.CurrentQuestionIndex-1], answers)
		result.Recap, result.RecapRound = &recap, q.CurrentQuestionIndex
	}
	return result, nil
}

func playerAnswers(board LeaderBoard, playerID uint64) []QuestionCorrectness {
	id := strconv.FormatUint(playerID, 10)
	for _, player := range board.PlayersPoint {
		if player.PlayerId == id {
			return player.QuestionCorrectness
		}
	}
	return nil
}

func (svc Service) writeEvent(conn net.Conn, code websocket.OpCode, event ProcessGameMessageResponse) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return svc.webSocket.WriteServerData(conn, code, string(data))
}
