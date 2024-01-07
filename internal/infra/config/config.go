package config

import "google.golang.org/grpc/keepalive"

type HttpServerConfig struct {
	Port uint
}

type DatabaseConfig struct {
	Driver                 string
	Url                    string
	ConnMaxLifeTimeMinutes int
	MaxOpenCons            int
	MaxIdleCons            int
}

type GrpcServerConfig struct {
	Port            uint32
	KeepaliveParams keepalive.ServerParameters
	KeepalivePolicy keepalive.EnforcementPolicy
}
