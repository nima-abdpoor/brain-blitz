package repository

type gameStatus struct {
	ExpectedNumberOfPlayers int   `json:"expectedNumberOfPlayers"`
	Players                 []int `json:"players"`
}
