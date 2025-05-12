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

type GameStatus uint8

const (
	GameStatusCreated GameStatus = iota + 1
	GameStatusPending
	GameStatusStarted
	GameStatusFinished
)

const (
	GSCreated  = "created"
	GSPending  = "pending"
	GSStarted  = "started"
	GSFinished = "finished"
)

func GetGameStatus() []GameStatus {
	return []GameStatus{GameStatusCreated, GameStatusPending, GameStatusStarted, GameStatusFinished}
}

func MapToGameStatus(status string) GameStatus {
	switch status {
	case GSCreated:
		return GameStatusCreated
	case GSPending:
		return GameStatusPending
	case GSStarted:
		return GameStatusStarted
	case GSFinished:
		return GameStatusFinished
	default:
		return 0
	}
}

func MapToFromGameStatus(status GameStatus) string {
	switch status {
	case GameStatusCreated:
		return GSCreated
	case GameStatusPending:
		return GSPending
	case GameStatusStarted:
		return GSStarted
	case GameStatusFinished:
		return GSFinished
	default:
		return "UNKNOWN"
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
