package grpc

import (
	pb "BrainBlitz.com/game/contract/auth/golang"
	"BrainBlitz.com/game/pkg/grpc"
	"fmt"
	"log/slog"
	"net"
)

type Server struct {
	server  grpc.RPCServer
	handler Handler
	logger  *slog.Logger
}

func NewServer(server grpc.RPCServer, handler Handler, logger *slog.Logger) Server {
	return Server{
		server:  server,
		handler: handler,
		logger:  logger,
	}
}

func (s Server) Serve() error {
	listener, err := net.Listen(s.server.Config.NetworkType, fmt.Sprintf(":%d", s.server.Config.Port))
	if err != nil {
		return err
	}
	pb.RegisterTokenServiceServer(s.server.Server, s.handler)
	if err := s.server.Server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func (s Server) Stop() {
	s.server.Stop()
}
