package entity

import "fmt"

type MatchedUsers struct {
	Category Category
	UserId   []uint
}

func (m MatchedUsers) String() string {
	return fmt.Sprintf("matchedUsers: Category: %s==>%s", MapFromCategory(m.Category), fmt.Sprint(m.UserId))
}
