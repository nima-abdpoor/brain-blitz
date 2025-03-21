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

func (svc Server) Serve() error {
	svc.RegisterRoutes()
	if err := svc.HTTPServer.Start(); err != nil {
		return err
	}
	return nil
}

func (svc Server) Stop(ctx context.Context) error {
	return svc.HTTPServer.Stop(ctx)
}

func (svc Server) RegisterRoutes() {
	v1 := svc.HTTPServer.Router.Group("/api/v1")
	v1.GET("/health-check", svc.healthCheck)
	v1.POST("/addToWaitingList", svc.Handler.addToWaitingList)
}
