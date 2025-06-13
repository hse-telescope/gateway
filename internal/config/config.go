package config

import (
	"os"

	"github.com/hse-telescope/logger"
	"gopkg.in/yaml.v3"
)

type Client struct {
	URL string `yaml:"url"`
}

type Config struct {
	Port      uint16 `yaml:"port"`
	PublicKey string `yaml:"public_key"`
	Clients   struct {
		Auth Client `yaml:"auth"`
		Core Client `yaml:"core"`
	} `yaml:"clients"`
	Logger           logger.Config `yaml:"logger"`
	OTELCollectorURL string        `yaml:"otel_collector_url"`
}

// Parse ...
func Parse(path string) (Config, error) {
	bytes, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return Config{}, err
	}

	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
