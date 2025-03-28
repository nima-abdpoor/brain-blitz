package http

import (
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"context"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	logger     logger.SlogAdapter
}

func New(server httpserver.Server, handler Handler, logger logger.SlogAdapter) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		logger:     logger,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}
	return nil
}

func (s Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) RegisterRoutes() {
	v1 := s.HTTPServer.Router.Group("/api/v1")
	v1.GET("/health-check", s.healthCheck)
	v1.POST("/access-token", s.Handler.CreateAccessToken)
	v1.POST("/refresh-token", s.Handler.CreateRefreshToken)
	v1.POST("/validate-token", s.Handler.ValidateToken)
	v1.GET("/validate-token", s.Handler.ValidateToken)
}
