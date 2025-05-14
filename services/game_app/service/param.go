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

type GameInitResponse struct {
	Categories      []string `json:"categories"`
	NumberOfPlayers []int    `json:"numberOfPlayers"`
}

type ProcessGameMessageRequest struct {
	MatchId         string  `json:"matchId"`
	Command         Command `json:"command"`
	Category        string  `json:"category"`
	NumberOfPlayers int     `json:"players"`
}

type ProcessGameMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
