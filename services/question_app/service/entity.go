package service

type Category string
type Difficulty string

const (
	CategoryTypeSport Category = "SPORT"
	CategoryTypeMusic Category = "MUSIC"
	CategoryTypeTech  Category = "TECH"
)

const (
	DifficultEasy   Difficulty = "EASY"
	DifficultMedium Difficulty = "MEDIUM"
	DifficultHard   Difficulty = "HARD"
)

func MapToCategory(category string) Category {
	switch category {
	case "MUSIC":
		return CategoryTypeMusic
	case "SPORT":
		return CategoryTypeSport
	case "TECH":
		return CategoryTypeTech
	default:
		return CategoryTypeMusic
	}
}

type MatchedUsers struct {
	Category []Category
	UserId   []uint64
}

type Question struct {
	Id            string
	Content       string
	CorrectAnswer string
	Choices       []string
	Category      Category
	Difficulty    Difficulty
}
