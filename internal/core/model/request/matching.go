package request

type AddToWaitingListRequest struct {
	Category string `json:"category"`
	UserId   int64
}
