package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Secret    string    `yaml:"secret"`
	BindTo    string    `yaml:"bind-to"`
	Domain    string    `yaml:"domain"`
	TLS       TLSConfig `yaml:"tls"`
	Blocklist string    `yaml:"blocklist"`
	SOCKS5    string    `yaml:"socks5"`
}

type TLSConfig struct {
	CertFile string `yaml:"cert-file"`
	KeyFile  string `yaml:"key-file"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}

	if cfg.BindTo == "" {
		return nil, fmt.Errorf("bind-to is required")
	}

	return &cfg, nil
}
