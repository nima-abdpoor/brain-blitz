package service

import (
	"fmt"
)

type Category string

const (
	CategoryTypeSport   Category = "SPORT"
	CategoryTypeMusic   Category = "MUSIC"
	CategoryTypeTech    Category = "TECH"
	CategoryTypeUnknown Category = "UNKNOWN"
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

type WaitingMember struct {
	UserId    uint
	TimeStamp int64
	Category  Category
}

func (wm WaitingMember) String() string {
	return fmt.Sprintf("UserId: %d, TimeStamp: %d, Category: %s", wm.UserId, wm.TimeStamp, wm.Category)
}

type MatchedUsers struct {
	Id       string
	Category []Category
	UserId   []uint64
}

func (m MatchedUsers) String() string {
	categories := ""
	for _, category := range m.Category {
		categories += fmt.Sprintf("%s,", string(category))
	}
	return fmt.Sprintf("matchedUsers: Category: %s==>%s", categories[:len(categories)-1], fmt.Sprint(m.UserId))
}
