package grpc

import (
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/user_app/service"
	"context"
)

type Handler struct {
	UserService service.Service
	Logger      logger.SlogAdapter
}

func NewHandler(srv service.Service, logger logger.SlogAdapter) Handler {
	return Handler{
		UserService: srv,
		Logger:      logger,
	}
}

func (s Handler) GetCustomer(ctx context.Context) error {
	return nil
}
