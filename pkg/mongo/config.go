package mongo

import "time"

type Config struct {
	User              string        `koanf:"user"`
	Host              string        `koanf:"host"`
	Port              int           `koanf:"port"`
	Name              string        `koanf:"name"`
	ConnectTimeout    time.Duration `koanf:"connect_timeout"`
	DisconnectTimeout time.Duration `koanf:"disconnect_timeout"`
}
