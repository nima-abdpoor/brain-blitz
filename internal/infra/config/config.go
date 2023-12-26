package config

type DatabaseConfig struct {
	Driver                 string
	Url                    string
	ConnMaxLifeTimeMinutes int
	MaxOpenCons            int
	MaxIdleCons            int
}
