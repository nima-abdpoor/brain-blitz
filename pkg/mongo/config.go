package mongo

import "time"

type Config struct {
	User              string        `koanf:"user"`
	Hosts             []string      `koanf:"host"`
	Ports             []int         `koanf:"port"`
	Name              string        `koanf:"name"`
	ConnectTimeout    time.Duration `koanf:"connect_timeout"`
	DisconnectTimeout time.Duration `koanf:"disconnect_timeout"`
	ReplicationName   string        `koanf:"replication_name"`
}
