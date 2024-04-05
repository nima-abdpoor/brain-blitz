package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log"
	"strings"
)

func Load(configPath string) *Config {
	var k = koanf.New(".")

	// Load default values using the confmap provider.
	// We provide a flat map with the "." delimiter.
	// A nested map can be loaded by setting the delimiter to an empty string "".
	err := k.Load(confmap.Provider(defaultConfig, "."), nil)
	if err != nil {
		log.Fatalf("Cant load Config from defaults %v", err)
	}

	// Load YAML config and merge into the previously loaded config (because we can).
	err = k.Load(file.Provider(configPath), yaml.Parser())
	if err != nil {
		log.Fatalf("Cant load Config from yaml file %v", err)
	}

	err = k.Load(env.Provider("BrainBlitz_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "BrainBlitz_")), "_", ".", -1)
	}), nil)
	if err != nil {
		log.Fatalf("Cant load Config from env %v", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
