package grpc

import (
	"BrainBlitz.com/game/user_app/service"
	"context"
	"log/slog"
)

type Handler struct {
	UserService service.Service
	Logger      *slog.Logger
}

func NewHandler(srv service.Service, logger *slog.Logger) Handler {
	return Handler{
		UserService: srv,
		Logger:      logger,
	}
}

func (s Handler) GetCustomer(ctx context.Context) error {
	return nil
}
