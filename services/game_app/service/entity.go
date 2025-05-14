package service

import (
	"time"
)

type Game struct {
	ID          uint
	PlayerIDs   []uint64
	QuestionIDs []uint
	Category    []Category
	Status      GameStatus
	StartTime   time.Time
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
	Id            string     `json:"id"`
	Content       string     `json:"content"`
	CorrectAnswer string     `json:"correctAnswer"`
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

type MatchCreation struct {
	Players  []uint64
	Category []string
	Status   string
}

type MatchedUsers struct {
	MatchId  string
	Category []Category
	UserId   []uint64
}
