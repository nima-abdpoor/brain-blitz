package entity

import "strconv"

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
