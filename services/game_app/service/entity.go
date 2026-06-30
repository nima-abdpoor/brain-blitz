package service

import (
	"time"
)

type Game struct {
	Id        *string     `bson:"id"`
	Players   []uint64    `bson:"players"`
	MatchId   string      `bson:"match_id"`
	Category  []Category  `bson:"category"`
	Status    GameStatus  `bson:"status"`
	Question  *[]Question `bson:"questions"`
	CreatedAt time.Time   `bson:"created_at"`
	UpdatedAt time.Time   `bson:"updated_at"`
}

type PlayerAnswer struct {
	GameId            string        `bson:"game_id"`
	QuestionIDs       string        `bson:"question_id"`
	PlayerID          string        `bson:"player_id"`
	PlayerChoice      string        `bson:"player_choice"`
	CorrectChoice     string        `bson:"correct_choice"`
	AnswerTime        time.Time     `bson:"answer_time"`
	ValidTimeToAnswer time.Time     `bson:"valid_time_to_answer"`
	TimeDiff          time.Duration `bson:"time_diff"`
	Options           []string      `bson:"Option"`
	Point             int           `bson:"point"`
	Category          Category      `bson:"category"`
}

type LeaderBoard struct {
	GameId       string
	PlayersPoint []PlayerPoint
}

type QuestionCorrectness struct {
	QuestionId    string
	PlayerChoice  string
	CorrectChoice string
	IsCorrect     bool
}

type PlayerPoint struct {
	PlayerId            string
	Point               int
	QuestionCorrectness []QuestionCorrectness
}

type GameStatus string

const (
	GameStatusInitialized GameStatus = "INITIALIZED"
	GameStatusPending     GameStatus = "PENDING"
	GameStatusCreated     GameStatus = "CREATED"
	GameStatusStarted     GameStatus = "STARTED"
	GameStatusFinished    GameStatus = "FINISHED"
	GameStatusUnknown     GameStatus = "UNKNOWN"
)

func GetGameStatus() []GameStatus {
	return []GameStatus{GameStatusCreated, GameStatusPending, GameStatusStarted, GameStatusFinished}
}

func MapToGameStatus(status string) GameStatus {
	switch status {
	case "PENDING":
		return GameStatusPending
	case "CREATED":
		return GameStatusCreated
	case "STARTED":
		return GameStatusStarted
	case "FINISHED":
		return GameStatusFinished
	default:
		return GameStatusUnknown
	}
}

type GameQuestions struct {
	Questions            []Question `json:"questions"`
	Players              []uint64   `json:"players"`
	CurrentQuestionIndex int        `bson:"currentQuestionIndex"`
}

type Question struct {
	Id              string     `json:"id"`
	Content         string     `json:"content"`
	CorrectAnswer   string     `json:"correctAnswer"`
	Status          string     `json:"status"`
	Choices         []string   `json:"choices"`
	Category        Category   `json:"category"`
	ValidAnswerTime time.Time  `json:"validAnswerTime"`
	Difficulty      Difficulty `json:"difficulty"`
}

type PossibleAnswers struct {
	ID     uint
	Text   string
	Choice PossibleAnswerChoice
}

type PossibleAnswerChoice uint8

type QuestionDifficulty uint8

const (
	PossibleAnswerChoiceA PossibleAnswerChoice = iota + 1
	PossibleAnswerChoiceB
	PossibleAnswerChoiceC
	PossibleAnswerChoiceD
)

type Category string
type Difficulty string

const (
	CategoryTypeSport   Category = "SPORT"
	CategoryTypeMusic   Category = "MUSIC"
	CategoryTypeTech    Category = "TECH"
	CategoryTypeUnknown Category = "UNKNOWN"
)

const (
	DifficultEasy    Difficulty = "EASY"
	DifficultMedium  Difficulty = "MEDIUM"
	DifficultHard    Difficulty = "HARD"
	DifficultUnknown Difficulty = "UNKNOWN"
)

func GetCategories() []Category {
	return []Category{CategoryTypeSport, CategoryTypeMusic, CategoryTypeTech}
}

func MapToCategory(category string) Category {
	switch category {
	case "SPORT":
		return CategoryTypeSport
	case "MUSIC":
		return CategoryTypeMusic
	case "TECH":
		return CategoryTypeTech
	default:
		return CategoryTypeUnknown
	}
}

func MapToDifficulty(difficulty string) Difficulty {
	switch difficulty {
	case "EASY":
		return DifficultEasy
	case "MEDIUM":
		return DifficultMedium
	case "HARD":
		return DifficultHard
	default:
		return DifficultUnknown
	}
}

func MapFromCategories(categories []Category) []string {
	var result []string
	for _, category := range categories {
		result = append(result, string(category))
	}
	return result
}

type GameQuestion struct {
	Id            string
	Text          string
	CorrectAnswer string
	Choices       []string
	Category      Category
}

type MatchedUsers struct {
	MatchId  string
	GameId   string
	Category []Category
	UserId   []uint64
}
