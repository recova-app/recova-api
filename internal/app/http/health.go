package http

import (
	"context"
	"strings"
	"time"
)

type dependencyMode string

const (
	// ReadinessModeRequired marks dependency as blocking for readiness success.
	ReadinessModeRequired dependencyMode = "required"
	// ReadinessModePlaceholder marks dependency as non-blocking placeholder.
	ReadinessModePlaceholder dependencyMode = "placeholder"
)

// ReadinessCheck defines one readiness dependency probe.
type ReadinessCheck struct {
	Name    string
	Mode    dependencyMode
	Message string
	Probe   func(ctx context.Context) error
}

func defaultReadinessChecks() []ReadinessCheck {
	return []ReadinessCheck{
		{
			Name:    "database",
			Mode:    ReadinessModePlaceholder,
			Message: "Probe database belum aktif; akan diikat saat koneksi database siap.",
			Probe: func(_ context.Context) error {
				return nil
			},
		},
	}
}

func (s *Server) evaluateReadiness(parent context.Context, requestIDs ...string) (map[string]any, bool) {
	reqID := ""
	if len(requestIDs) > 0 {
		reqID = strings.TrimSpace(requestIDs[0])
	}

	timeout := s.readinessTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	checkCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	ready := true
	checks := make(map[string]any, len(s.readinessChecks))

	for _, check := range s.readinessChecks {
		name := strings.TrimSpace(check.Name)
		if name == "" {
			continue
		}

		err := check.Probe(checkCtx)
		status := "ok"
		message := strings.TrimSpace(check.Message)

		if check.Mode == ReadinessModePlaceholder {
			status = "placeholder"
		} else if err != nil {
			status = "down"
			ready = false
			if message == "" {
				message = "Dependency tidak sehat"
			}
			if s.logger != nil {
				s.logger.Error("readiness probe failed",
					"request_id", reqID,
					"dependency", name,
					"mode", string(check.Mode),
					"error", err,
				)
			}
		}

		checks[name] = map[string]any{
			"status":  status,
			"healthy": err == nil || check.Mode == ReadinessModePlaceholder,
			"message": message,
		}
	}

	return checks, ready
}
