package service

import (
	"fmt"
	"strconv"
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

type WaitingMember struct {
	UserId    uint
	TimeStamp int64
	Category  Category
}

func (wm WaitingMember) String() string {
	return fmt.Sprintf("UserId: %d, TimeStamp: %d, Category: %s", wm.UserId, wm.TimeStamp, wm.Category)
}

type MatchedUsers struct {
	Category []Category
	UserId   []uint64
}

func (m MatchedUsers) String() string {
	categories := ""
	for _, category := range m.Category {
		categories += fmt.Sprintf("%s,", MapFromCategory(category))
	}
	return fmt.Sprintf("matchedUsers: Category: %s==>%s", categories[:len(categories)-1], fmt.Sprint(m.UserId))
}
