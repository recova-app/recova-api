// Package config tests configuration loading and validation behavior.
package config

import "testing"

// TestLoad_RequiredEnvMissing_ReturnsError ensures missing required env fails fast.
func TestLoad_RequiredEnvMissing_ReturnsError(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("API_PREFIX", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when required env is missing")
	}
}

// TestLoad_RequiredEnvPresent_ReturnsConfig ensures valid required env can be loaded.
func TestLoad_RequiredEnvPresent_ReturnsConfig(t *testing.T) {
	t.Setenv("APP_NAME", "recova-backend")
	t.Setenv("APP_ENV", "local")
	t.Setenv("PORT", "3000")
	t.Setenv("API_PREFIX", "/api/v1")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.AppName != "recova-backend" {
		t.Fatalf("unexpected app name: %s", cfg.AppName)
	}
}
