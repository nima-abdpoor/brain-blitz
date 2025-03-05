package http

import "BrainBlitz.com/game/match_app/service"

type Handler struct {
	Service service.Service
}

func NewHandler(userService service.Service) Handler {
	return Handler{
		Service: userService,
	}
}
