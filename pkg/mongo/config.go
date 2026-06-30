package mongo

import "time"

type Instance struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type Config struct {
	Instances         []Instance    `koanf:"mongo_instances"`
	User              string        `koanf:"user"`
	Hosts             []string      `koanf:"host"`
	Ports             []string      `koanf:"port"`
	Name              string        `koanf:"name"`
	ConnectTimeout    time.Duration `koanf:"connect_timeout"`
	DisconnectTimeout time.Duration `koanf:"disconnect_timeout"`
	ReplicationName   string        `koanf:"replication_name"`
}
