package service

import "time"

type AddToWaitingListResponse struct {
	Timeout time.Duration `json:"timeout"`
}

type AddToWaitingListRequest struct {
	Category string `json:"category"`
	UserId   string
}

type MatchWaitedUsersRequest struct {
	Category string `json:"category"`
}

type MatchWaitedUsersResponse struct {
	WaitingUsers []string `json:"waitingUsers"`
}
