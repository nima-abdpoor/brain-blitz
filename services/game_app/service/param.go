package service

type ProcessGameRequest struct {
	Id string
}

type ProcessGameResponse struct {
}

type Command string
type Event string

const (
	CommandGetCategories    Command = "GET_CATEGORIES"
	CommandAddToWaitingList Command = "ADD_TO_WAITING_LIST"
	CommandAnswer           Command = "ANSWER"
	CommandReady            Command = "READY"
	CommandUnknownCommand   Command = "UNKNOWN"
)

const (
	EventMatchCreated Event = "MATCH_CREATED"
	NewQuestion       Event = "NEW_QUESTION"
	AnswerAccepted    Event = "ANSWER_ACCEPTED"
)

type GameInitResponse struct {
	Categories      []string `json:"categories"`
	NumberOfPlayers []int    `json:"numberOfPlayers"`
}

type ProcessGameMessageRequest struct {
	MatchId         string            `json:"matchId"`
	GameId          string            `json:"gameId"`
	Command         Command           `json:"command"`
	Category        string            `json:"category"`
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
	GameId   string                 `json:"gameId"`
	Question ProcessGameNewQuestion `json:"question"`
	Answer   ProcessGameLeaderBoard `json:"leaderBoard"`
}

type ProcessGameLeaderBoard struct {
	GameId      string                   `json:"gameId"`
	PlayerPoint []ProcessGamePlayerPoint `json:"playerPoint"`
}

type ProcessGamePlayerPoint struct {
	PlayerId string                    `json:"playerId"`
	Point    int                       `json:"point"`
	Answers  []ProcessGameAnswerResult `json:"answers"`
}

type ProcessGameAnswerResult struct {
	QuestionId    string `json:"questionId"`
	CorrectAnswer string `json:"correctAnswer"`
	PlayerAnswer  string `json:"playerAnswer"`
	IsCorrect     bool   `json:"isCorrect"`
}

type ProcessGameNewQuestion struct {
	Id         string     `json:"id"`
	Content    string     `json:"content"`
	Choices    []string   `json:"choices"`
	Difficulty Difficulty `json:"difficulty"`
}
