package repository

import entity "BrainBlitz.com/game/entity/game"

type MatchMakingRepository interface {
	AddToWaitingList(category entity.Category, userId string) error
}
