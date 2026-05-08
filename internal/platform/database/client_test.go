package database

import (
	"testing"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

func TestClientClose_Nil_NoError(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Fatalf("expected nil close error, got %v", err)
	}
}

func TestClientPing_NilClient_ReturnsError(t *testing.T) {
	var c *Client
	if err := c.Ping(nil); err == nil {
		t.Fatal("expected ping error for nil client")
	}
}

func TestConnect_UnreachableDatabase_ReturnsError(t *testing.T) {
	cfg := config.Config{
		Database: config.DatabaseConfig{
			URL:                "postgresql://invalid:invalid@127.0.0.1:1/recova_db?sslmode=disable",
			MaxOpenConns:       5,
			MaxIdleConns:       2,
			ConnMaxLifetimeSec: 60,
		},
		Observability: config.ObservabilityConfig{
			HealthCheckTimeoutMs: 50,
		},
	}

	client, err := Connect(cfg)
	if err == nil {
		_ = client.Close()
		t.Fatal("expected connection error for unreachable database")
	}
}
