package feature

type Config struct {
	Infra   bool `kafka:"infra"`
	Metrics bool `koanf:"metrics"`
	PPROF   bool `koanf:"pprof"`
}
