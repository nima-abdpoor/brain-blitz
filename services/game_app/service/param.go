package service

import "time"

type ProcessGameRequest struct {
	Id string
}

type ProcessGameResponse struct {
}

type Command string
type Event string

const (
	CommandReady            Command = "READY"
	CommandAnswer           Command = "ANSWER"
	CommandUnknownCommand   Command = "UNKNOWN"
	CommandGetCategories    Command = "GET_CATEGORIES"
	CommandAddToWaitingList Command = "ADD_TO_WAITING_LIST"
)

const (
	Error              Event = "ERROR"
	EventMatchCreated  Event = "MATCH_CREATED"
	EventRoundClosed   Event = "ROUND_CLOSED"
	AnswerAccepted     Event = "ANSWER_ACCEPTED"
	NewQuestion        Event = "QUESTIONS_PUBLISHED"
	AddedToWaitingList Event = "ADDED_TO_WAITING_LIST"
	EventCompleted     Event = "COMPLETED"
)

type GameInitResponse struct {
	Categories      []string `json:"categories"`
	NumberOfPlayers []int    `json:"numberOfPlayers"`
}

type ProcessGameMessageRequest struct {
	TimeOptionID    string            `json:"timeOptionId,omitempty"`
	MatchId         string            `json:"matchId"`
	GameId          string            `json:"gameId"`
	Command         Command           `json:"command"`
	Category        string            `json:"category"`
	DisplayName     string            `json:"displayName,omitempty"`
	NumberOfPlayers int               `json:"players"`
	GameAnswer      ProcessGameAnswer `json:"answer"`
}

type ProcessGameAnswer struct {
	GameId     string `json:"gameId"`
	QuestionId string `json:"questionId"`
	Answer     string `json:"choice"`
}

type ProcessGameMessageResponse struct {
	Success  bool                        `json:"success"`
	Event    Event                       `json:"event"`
	Message  string                      `json:"message"`
	MetaData ProcessGameMetaDataResponse `json:"metaData"`
}

type ProcessGameMetaDataResponse struct {
	Recap          *RoundRecap                `json:"recap,omitempty"`
	AnswerFeedback *ProcessGameAnswerFeedback `json:"answerFeedback,omitempty"`
	RoundNumber    int                        `json:"roundNumber,omitempty"`
	TotalQuestions int                        `json:"totalQuestions,omitempty"`
	GameId         string                     `json:"gameId"`
	Questions      []ProcessGameQuestion      `json:"questions"`
	Answer         ProcessGameLeaderBoard     `json:"leaderBoard"`
}

// ProcessGameAnswerFeedback reports only the submitting player's verdict.
// The correct answer is revealed separately when the round closes.
type ProcessGameAnswerFeedback struct {
	QuestionId string `json:"questionId"`
	IsCorrect  bool   `json:"isCorrect"`
}

type ProcessGameLeaderBoard struct {
	GameId      string                   `json:"gameId"`
	PlayerPoint []ProcessGamePlayerPoint `json:"playerPoint"`
}

type ProcessGamePlayerPoint struct {
	DisplayName string                    `json:"displayName,omitempty"`
	PlayerId    string                    `json:"playerId"`
	Point       int                       `json:"point"`
	Answers     []ProcessGameAnswerResult `json:"answers"`
}

type ProcessGameAnswerResult struct {
	QuestionId    string `json:"questionId"`
	CorrectAnswer string `json:"correctAnswer"`
	PlayerAnswer  string `json:"playerAnswer"`
	IsCorrect     bool   `json:"isCorrect"`
}

type ProcessGameQuestion struct {
	Id         string     `json:"id"`
	Content    string     `json:"content"`
	Choices    []string   `json:"choices"`
	Difficulty Difficulty `json:"difficulty"`
	TTL        time.Time  `json:"ttl"`
}
