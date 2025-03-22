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
	public := s.HTTPServer.Router.Group("/public/api/v1")
	private := s.HTTPServer.Router.Group("/api/v1")
	public.GET("/health-check", s.healthCheck)
	public.POST("/signup", s.Handler.SignUp)
	public.POST("/login", s.Handler.Login)
	private.GET("/profile", s.Handler.Profile)
}
