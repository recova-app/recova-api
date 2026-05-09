// Package observability contains metrics, tracing, and diagnostics adapters.
package observability

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	unknownLabelValue = "unknown"
	statusSuccess     = "success"
	statusFailure     = "failure"
)

// Recorder stores and exports runtime observability metrics.
type Recorder struct {
	registry *prometheus.Registry

	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDurationSec *prometheus.HistogramVec
	httpErrorsTotal        *prometheus.CounterVec

	dbOperationDurationSec *prometheus.HistogramVec
	aiRequestDurationSec   *prometheus.HistogramVec
	aiPersonaUsageTotal    *prometheus.CounterVec

	auditEventsTotal *prometheus.CounterVec
}

// NewRecorder creates an isolated metrics recorder with a dedicated registry.
func NewRecorder() *Recorder {
	reg := prometheus.NewRegistry()

	recorder := &Recorder{
		registry: reg,
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recova_http_requests_total",
			Help: "Total HTTP requests handled by route and status class.",
		}, []string{"method", "route", "status_class"}),
		httpRequestDurationSec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recova_http_request_duration_seconds",
			Help:    "HTTP request latency by route and status class.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route", "status_class"}),
		httpErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recova_http_errors_total",
			Help: "Total HTTP errors by route and error class.",
		}, []string{"method", "route", "error_class"}),
		dbOperationDurationSec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recova_db_operation_duration_seconds",
			Help:    "Database operation latency by operation/table/status.",
			Buckets: prometheus.DefBuckets,
		}, []string{"operation", "table", "status"}),
		aiRequestDurationSec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recova_ai_request_duration_seconds",
			Help:    "AI provider request latency by provider/model/status.",
			Buckets: prometheus.DefBuckets,
		}, []string{"provider", "model", "status"}),
		aiPersonaUsageTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recova_ai_persona_usage_total",
			Help: "Total AI persona usage events by action, persona, and status.",
		}, []string{"action", "persona", "status"}),
		auditEventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recova_audit_events_total",
			Help: "Total audit events by action and result.",
		}, []string{"action", "result"}),
	}

	reg.MustRegister(
		recorder.httpRequestsTotal,
		recorder.httpRequestDurationSec,
		recorder.httpErrorsTotal,
		recorder.dbOperationDurationSec,
		recorder.aiRequestDurationSec,
		recorder.aiPersonaUsageTotal,
		recorder.auditEventsTotal,
	)

	return recorder
}

// MetricsHandler returns HTTP handler exposing recorder metrics in Prometheus format.
func (r *Recorder) MetricsHandler() http.Handler {
	if r == nil || r.registry == nil {
		return promhttp.Handler()
	}
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest captures request count/latency and error counters.
func (r *Recorder) RecordHTTPRequest(method string, route string, statusCode int, duration time.Duration) {
	if r == nil {
		return
	}

	normalizedMethod := normalizeMethod(method)
	normalizedRoute := normalizeRoute(route)
	class := httpStatusClass(statusCode)

	r.httpRequestsTotal.WithLabelValues(normalizedMethod, normalizedRoute, class).Inc()
	r.httpRequestDurationSec.WithLabelValues(normalizedMethod, normalizedRoute, class).Observe(duration.Seconds())
	if statusCode >= 400 {
		r.httpErrorsTotal.WithLabelValues(normalizedMethod, normalizedRoute, class).Inc()
	}
}

// RecordDBOperation captures database operation latency and status.
func (r *Recorder) RecordDBOperation(operation string, table string, duration time.Duration, err error) {
	if r == nil {
		return
	}
	status := statusSuccess
	if err != nil {
		status = statusFailure
	}
	r.dbOperationDurationSec.WithLabelValues(
		normalizeLabel(operation),
		normalizeLabel(table),
		status,
	).Observe(duration.Seconds())
}

// RecordAIRequest captures AI provider latency and status.
func (r *Recorder) RecordAIRequest(provider string, model string, duration time.Duration, err error) {
	if r == nil {
		return
	}
	status := statusSuccess
	if err != nil {
		status = statusFailure
	}
	r.aiRequestDurationSec.WithLabelValues(
		normalizeLabel(provider),
		normalizeLabel(model),
		status,
	).Observe(duration.Seconds())
}

// RecordAIPersonaUsage captures persona usage distribution and outcome status.
func (r *Recorder) RecordAIPersonaUsage(action string, persona string, err error) {
	if r == nil {
		return
	}
	status := statusSuccess
	if err != nil {
		status = statusFailure
	}
	r.aiPersonaUsageTotal.WithLabelValues(
		normalizeLabel(action),
		normalizeLabel(persona),
		status,
	).Inc()
}

// RecordAuditEvent increments audit event counter for one action and result.
func (r *Recorder) RecordAuditEvent(action string, result string) {
	if r == nil {
		return
	}
	r.auditEventsTotal.WithLabelValues(
		normalizeLabel(action),
		normalizeLabel(result),
	).Inc()
}

func normalizeMethod(method string) string {
	normalized := strings.ToUpper(strings.TrimSpace(method))
	if normalized == "" {
		return unknownLabelValue
	}
	return normalized
}

func normalizeRoute(route string) string {
	trimmed := strings.TrimSpace(route)
	if trimmed == "" {
		return unknownLabelValue
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func normalizeLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return unknownLabelValue
	}
	return trimmed
}

func httpStatusClass(statusCode int) string {
	if statusCode < 100 || statusCode > 999 {
		return unknownLabelValue
	}
	prefix := statusCode / 100
	return strconv.Itoa(prefix) + "xx"
}
