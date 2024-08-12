package config

import "google.golang.org/grpc/keepalive"

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

type Infra struct {
	PPROF bool `koanf:"pprof"`
}
