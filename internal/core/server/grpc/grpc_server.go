package grpc

import (
	"BrainBlitz.com/game/internal/infra/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"time"
)

type GRPCServer interface {
	Start(serviceRegister func(server *grpc.Server))
	io.Closer
}

type gRPCServer struct {
	grpcServer *grpc.Server
	config     config.GrpcServerConfig
}

func (g gRPCServer) Start(serviceRegister func(server *grpc.Server)) {
	//TODO implement me
	panic("implement me")
}

func (g gRPCServer) Close() error {
	//TODO implement me
	panic("implement me")
}

func NewGRPCServer(config config.GrpcServerConfig) (GRPCServer, error) {
	options, err := buildOptions(config)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(options...)
	return &gRPCServer{
		grpcServer: server,
		config:     config,
	}, err
}

func buildOptions(config config.GrpcServerConfig) ([]grpc.ServerOption, error) {
	return []grpc.ServerOption{
		grpc.KeepaliveParams(buildKeepaliveParams(config.KeepaliveParams)),
		grpc.KeepaliveEnforcementPolicy(buildKeepalivePolicy(config.KeepalivePolicy)),
	}, nil
}

func buildKeepalivePolicy(config keepalive.EnforcementPolicy) keepalive.EnforcementPolicy {
	return keepalive.EnforcementPolicy{
		MinTime:             config.MinTime * time.Second,
		PermitWithoutStream: config.PermitWithoutStream,
	}
}

func buildKeepaliveParams(config keepalive.ServerParameters) keepalive.ServerParameters {
	return keepalive.ServerParameters{
		MaxConnectionIdle:     config.MaxConnectionIdle * time.Second,
		MaxConnectionAge:      config.MaxConnectionAge * time.Second,
		MaxConnectionAgeGrace: config.MaxConnectionAgeGrace * time.Second,
		Time:                  config.Time * time.Second,
		Timeout:               config.Timeout * time.Second,
	}
}
