package entity

type Category struct {
	ID          uint
	Type        CategoryType
	Description string
}

type CategoryType uint8

const (
	CategoryTypeSport CategoryType = iota + 1
	CategoryTypeMusic
	CategoryTypeTech
)
