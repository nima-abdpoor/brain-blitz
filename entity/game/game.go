package entity

import "time"

type Game struct {
	ID          uint
	PlayerIDs   []uint
	QuestionIDs []uint
	Category    []uint
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
