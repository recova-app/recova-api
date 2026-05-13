package lifecycle

import (
	"context"
	"testing"
)

func TestWithShutdownSignal_CancelStopsContext(t *testing.T) {
	ctx, cancel := WithShutdownSignal(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		// ok
	default:
		t.Fatal("expected ctx done after cancel")
	}
}
