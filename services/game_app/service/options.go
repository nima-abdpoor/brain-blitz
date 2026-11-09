package service

type GameOptions struct {
	MinPlayers      int   `json:"minPlayers"`
	MaxPlayers      int   `json:"maxPlayers"`
	NumberOfPlayers []int `json:"numberOfPlayers"`
}

func (svc Service) GameOptions() GameOptions {
	return GameOptions{MinPlayers: svc.config.Players.Min, MaxPlayers: svc.config.Players.Max, NumberOfPlayers: svc.config.Players.Options()}
}
