package http

import (
	"context"
	"strings"
	"time"
)

type dependencyMode string

const (
	dependencyModeRequired    dependencyMode = "required"
	dependencyModePlaceholder dependencyMode = "placeholder"
)

type dependencyCheck struct {
	Name    string
	Mode    dependencyMode
	Message string
	Probe   func(ctx context.Context) error
}

func defaultReadinessChecks() []dependencyCheck {
	return []dependencyCheck{
		{
			Name:    "database",
			Mode:    dependencyModePlaceholder,
			Message: "Probe database belum aktif; akan diikat saat koneksi database siap.",
			Probe: func(_ context.Context) error {
				return nil
			},
		},
	}
}

func (s *Server) evaluateReadiness(parent context.Context) (map[string]any, bool) {
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

		switch {
		case check.Mode == dependencyModePlaceholder:
			status = "placeholder"
		case err != nil:
			status = "down"
			ready = false
			if message == "" {
				message = "Dependency tidak sehat"
			}
		}

		if err != nil {
			if message == "" {
				message = err.Error()
			} else {
				message = message + " (" + err.Error() + ")"
			}
		}

		checks[name] = map[string]any{
			"status":  status,
			"healthy": err == nil || check.Mode == dependencyModePlaceholder,
			"message": message,
		}
	}

	return checks, ready
}
