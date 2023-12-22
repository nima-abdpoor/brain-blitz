package entity

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
