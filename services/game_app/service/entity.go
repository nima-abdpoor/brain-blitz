package service

import (
	"time"
)

type Game struct {
	Id                   *string     `bson:"id"`
	Players              []uint64    `bson:"players"`
	MatchId              string      `bson:"match_id"`
	Category             []Category  `bson:"category"`
	Status               GameStatus  `bson:"status"`
	CurrentQuestionIndex int         `bson:"current_question_index"`
	Question             *[]Question `bson:"questions"`
	CreatedAt            time.Time   `bson:"created_at"`
	UpdatedAt            time.Time   `bson:"updated_at"`
}

type Player struct {
	ID     uint
	UserID uint
	GameID uint
	Score  int
	Answer []PlayerAnswer
}

type PlayerAnswer struct {
	ID          uint
	QuestionIDs []uint
	PlayerID    []uint
	Choice      PossibleAnswerChoice
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

type Question struct {
	Id            string `json:"id"`
	Content       string `json:"content"`
	CorrectAnswer string `json:"correctAnswer"`
	Status        string
	Choices       []string   `json:"choices"`
	Category      Category   `json:"category"`
	Difficulty    Difficulty `json:"difficulty"`
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
