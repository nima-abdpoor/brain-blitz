package request

type AddToWaitingListRequest struct {
	Category string `json:"category"`
	UserId   string
}

type MatchWaitedUsersRequest struct{}
