package entity

import "fmt"

type WaitingMember struct {
	UserId    uint
	TimeStamp int64
	Category  Category
}

func (wm WaitingMember) String() string {
	return fmt.Sprintf("UserId: %d, TimeStamp: %d, Category: %s", wm.UserId, wm.TimeStamp, wm.Category)
}
