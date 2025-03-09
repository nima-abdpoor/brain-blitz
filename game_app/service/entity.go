package service

import (
	"strconv"
	"time"
)

type Game struct {
	ID          uint
	PlayerIDs   []uint64
	QuestionIDs []uint
	Category    Category
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
	ID              uint
	Text            string
	PossibleAnswers []PossibleAnswers
	CorrectAnswerID uint
	Difficulty      QuestionDifficulty
	CategoryID      uint
}

type PossibleAnswers struct {
	ID     uint
	Text   string
	Choice PossibleAnswerChoice
}

type PossibleAnswerChoice uint8

type QuestionDifficulty uint8

const (
	QuestionDifficultyEasy QuestionDifficulty = iota + 1
	QuestionDifficultyMedium
	QuestionDifficultyHard
)

const (
	PossibleAnswerChoiceA PossibleAnswerChoice = iota + 1
	PossibleAnswerChoiceB
	PossibleAnswerChoiceC
	PossibleAnswerChoiceD
)

type Category uint8

const (
	CategoryTypeSport Category = iota + 1
	CategoryTypeMusic
	CategoryTypeTech
)

const (
	Sport = "sport"
	Music = "music"
	Tech  = "technology"
)

func GetCategories() []Category {
	return []Category{CategoryTypeSport, CategoryTypeMusic, CategoryTypeTech}
}

func MapToCategory(category string) Category {
	switch category {
	case Music:
		return CategoryTypeMusic
	case Sport:
		return CategoryTypeSport
	case Tech:
		return CategoryTypeTech
	//todo select randomly
	default:
		return 0
	}
}

func MapFromCategory(category Category) string {
	switch category {
	case CategoryTypeMusic:
		return Music
	case CategoryTypeSport:
		return Sport
	case CategoryTypeTech:
		return Tech
	// todo select randomly
	default:
		return "Unknown"
	}
}

func (c Category) String() string {
	return strconv.Itoa(int(c))
}

type MatchCreation struct {
	Players  []uint64
	Category []string
	Status   string
}
