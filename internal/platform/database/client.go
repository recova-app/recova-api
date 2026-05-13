package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const defaultPingTimeout = 2 * time.Second

// Client wraps GORM and sql.DB handles for lifecycle and health checks.
type Client struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

// Connect opens PostgreSQL connection via GORM, applies pool configuration, and pings once.
func Connect(cfg config.Config) (*Client, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql database handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetimeSec) * time.Second)

	timeout := time.Duration(cfg.Observability.HealthCheckTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = defaultPingTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("database connection is unhealthy: %w", err)
	}

	return &Client{
		gormDB: gormDB,
		sqlDB:  sqlDB,
	}, nil
}

// Gorm returns the root GORM handle.
func (c *Client) Gorm() *gorm.DB {
	if c == nil {
		return nil
	}
	return c.gormDB
}

// Ping runs database connectivity check using the provided context.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.sqlDB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return c.sqlDB.PingContext(ctx)
}

// Close releases underlying sql.DB resources.
func (c *Client) Close() error {
	if c == nil || c.sqlDB == nil {
		return nil
	}
	return c.sqlDB.Close()
}
