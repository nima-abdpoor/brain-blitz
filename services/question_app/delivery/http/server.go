package http

import (
	httpserver "BrainBlitz.com/game/pkg/http_server"
	"BrainBlitz.com/game/pkg/logger"
	"context"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	logger     logger.Logger
}

func New(server httpserver.Server, handler Handler, logger logger.Logger) Server {
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
	v1 := s.HTTPServer.Router.Group("/api/v1/question")
	v1.GET("/health-check", s.healthCheck)
	v1.POST("/add", s.Handler.AddQuestion)
	v1.GET("/time-options", s.Handler.TimeOptions)
	v1.GET("/admin/time-options", s.Handler.AdminTimeOptions)
	v1.PUT("/admin/time-options/:id", s.Handler.PutTimeOption)
}
