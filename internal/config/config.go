package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Path string `yaml:"path"`
}

func LoadConfig() (*Config, error) {
	// Highest priority: ENV variable
	if envPath := os.Getenv("SEMVER_PATH"); envPath != "" {
		return &Config{Path: envPath}, nil
	}

	// Second priority: YAML file
	data, err := os.ReadFile(".semver.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // fallback to default
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Extra safety: ensure the `path` key was present
	if cfg.Path == "" {
		return nil, errors.New("invalid config: missing or empty 'path'")
	}

	return &cfg, nil
}
