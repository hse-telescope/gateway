package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Client struct {
	URL string `yaml:"url"`
}

type Config struct {
	Port    uint16   `yaml:"port"`
	Clients []Client `yaml:"clients"`
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
