package entity

type GameStatus uint8

const (
	GameStatusCreated GameStatus = iota + 1
	GameStatusPending
	GameStatusStarted
	GameStatusFinished
)

const (
	GSCreated  = "created"
	GSPending  = "pending"
	GSStarted  = "started"
	GSFinished = "finished"
)

func GetGameStatus() []GameStatus {
	return []GameStatus{GameStatusCreated, GameStatusPending, GameStatusStarted, GameStatusFinished}
}

func MapToGameStatus(status string) GameStatus {
	switch status {
	case GSCreated:
		return GameStatusCreated
	case GSPending:
		return GameStatusPending
	case GSStarted:
		return GameStatusStarted
	case GSFinished:
		return GameStatusFinished
	default:
		return 0
	}
}

func MapToFromGameStatus(status GameStatus) string {
	switch status {
	case GameStatusCreated:
		return GSCreated
	case GameStatusPending:
		return GSPending
	case GameStatusStarted:
		return GSStarted
	case GameStatusFinished:
		return GSFinished
	default:
		return "UNKNOWN"
	}
}
