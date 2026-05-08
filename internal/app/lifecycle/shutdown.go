// Package lifecycle manages top-level process lifecycle utilities.
package lifecycle

import (
	"context"
	"os/signal"
	"syscall"
)

// WithShutdownSignal returns a context canceled when a process shutdown signal is received.
func WithShutdownSignal(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
}
