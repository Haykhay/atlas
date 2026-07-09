// Package config reads and writes the Atlas configuration file.
// Credentials are never stored here — they live in the OS secure
// credential store (see internal/provider in Plan 2).
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProviderConfig struct {
	Model string `yaml:"model,omitempty"`
}

type Config struct {
	DefaultProvider string                    `yaml:"default_provider,omitempty"`
	Providers       map[string]ProviderConfig `yaml:"providers,omitempty"`
}

// Dir returns the Atlas config directory. ATLAS_CONFIG_DIR overrides the
// platform default (tests rely on this).
func Dir() (string, error) {
	if dir := os.Getenv("ATLAS_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "atlas"), nil
}

func filePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	p, err := filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Providers: map[string]ProviderConfig{}}, nil
	}
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Providers == nil {
		cfg.Providers = map[string]ProviderConfig{}
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	p, err := filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
