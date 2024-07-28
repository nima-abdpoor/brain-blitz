package entity

import "time"

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
