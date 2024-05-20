package response

import "time"

type AddToWaitingListResponse struct {
	Timeout time.Duration `json:"timeout"`
}

type MatchWaitedUsersResponse struct {
	WaitingUsers []string `json:"waitingUsers"`
}
