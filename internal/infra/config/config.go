package config

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
