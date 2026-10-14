package observability

import (
	"database/sql"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type dbPoolCollector struct {
	db *sql.DB

	openConnections     *prometheus.Desc
	inUseConnections    *prometheus.Desc
	idleConnections     *prometheus.Desc
	waitCount           *prometheus.Desc
	waitDurationSeconds *prometheus.Desc
	maxOpenConnections  *prometheus.Desc
}

func newDBPoolCollector(db *sql.DB) *dbPoolCollector {
	return &dbPoolCollector{
		db: db,
		openConnections: prometheus.NewDesc(
			"recova_db_pool_open_connections",
			"The number of established database connections both in use and idle.",
			nil,
			nil,
		),
		inUseConnections: prometheus.NewDesc(
			"recova_db_pool_in_use_connections",
			"The number of database connections currently in use.",
			nil,
			nil,
		),
		idleConnections: prometheus.NewDesc(
			"recova_db_pool_idle_connections",
			"The number of idle database connections.",
			nil,
			nil,
		),
		waitCount: prometheus.NewDesc(
			"recova_db_pool_wait_count_total",
			"The total number of database connections waited for.",
			nil,
			nil,
		),
		waitDurationSeconds: prometheus.NewDesc(
			"recova_db_pool_wait_duration_seconds_total",
			"The total time blocked waiting for a new database connection in seconds.",
			nil,
			nil,
		),
		maxOpenConnections: prometheus.NewDesc(
			"recova_db_pool_max_open_connections",
			"Maximum number of open connections to the database.",
			nil,
			nil,
		),
	}
}

// Describe sends all possible metric descriptors to the provided channel.
func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	if c == nil {
		return
	}
	ch <- c.openConnections
	ch <- c.inUseConnections
	ch <- c.idleConnections
	ch <- c.waitCount
	ch <- c.waitDurationSeconds
	ch <- c.maxOpenConnections
}

// Collect reads sql.DBStats and emits current gauge and counter values.
func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	if c == nil || c.db == nil {
		return
	}
	stats := c.db.Stats()

	ch <- prometheus.MustNewConstMetric(c.openConnections, prometheus.GaugeValue, float64(stats.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.inUseConnections, prometheus.GaugeValue, float64(stats.InUse))
	ch <- prometheus.MustNewConstMetric(c.idleConnections, prometheus.GaugeValue, float64(stats.Idle))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(stats.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.waitDurationSeconds, prometheus.CounterValue, stats.WaitDuration.Seconds())
	ch <- prometheus.MustNewConstMetric(c.maxOpenConnections, prometheus.GaugeValue, float64(stats.MaxOpenConnections))
}

// RegisterDatabasePoolMetrics registers database connection pool metrics with the recorder registry.
func RegisterDatabasePoolMetrics(sqlDB *sql.DB, recorder *Recorder) error {
	if sqlDB == nil || recorder == nil {
		return nil
	}

	customCollector := newDBPoolCollector(sqlDB)
	if err := recorder.RegisterCollector(customCollector); err != nil {
		return fmt.Errorf("failed to register db pool collector: %w", err)
	}

	goDBStats := collectors.NewDBStatsCollector(sqlDB, "recova")
	if err := recorder.RegisterCollector(goDBStats); err != nil {
		return fmt.Errorf("failed to register go db stats collector: %w", err)
	}

	return nil
}
