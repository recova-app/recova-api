// Package config provides runtime configuration loading and validation.
package config

import (
	"fmt"
	"os"
	"strings"
)

// AppConfig contains minimal runtime configuration for the API bootstrap stage.
type AppConfig struct {
	AppName   string
	AppEnv    string
	Port      string
	APIPrefix string
}

// Load reads and validates required configuration from environment variables.
func Load() (AppConfig, error) {
	cfg := AppConfig{
		AppName:   strings.TrimSpace(os.Getenv("APP_NAME")),
		AppEnv:    strings.TrimSpace(os.Getenv("APP_ENV")),
		Port:      strings.TrimSpace(os.Getenv("PORT")),
		APIPrefix: strings.TrimSpace(os.Getenv("API_PREFIX")),
	}

	if cfg.AppName == "" {
		return AppConfig{}, fmt.Errorf("APP_NAME wajib diisi")
	}

	if cfg.AppEnv == "" {
		return AppConfig{}, fmt.Errorf("APP_ENV wajib diisi")
	}

	if cfg.Port == "" {
		return AppConfig{}, fmt.Errorf("PORT wajib diisi")
	}

	if cfg.APIPrefix == "" {
		return AppConfig{}, fmt.Errorf("API_PREFIX wajib diisi")
	}

	return cfg, nil
}
