package service

type ProcessGameRequest struct {
	Id string
}

type ProcessGameResponse struct {
}

type Command string

const (
	CommandGetCategories    Command = "GET_CATEGORIES"
	CommandAddToWaitingList Command = "ADD_TO_WAITING_LIST"
	CommandReady            Command = "READY"
	CommandUnknownCommand   Command = "UNKNOWN"
)

type ProcessGameMessageRequest struct {
	MatchId string  `json:"matchId"`
	Command Command `json:"command"`
}
