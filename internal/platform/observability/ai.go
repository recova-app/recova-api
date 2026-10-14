package observability

import (
	"context"
	"strings"
	"time"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
)

type fallbackObservable interface {
	SetFallbackObserver(observer aiplatform.FallbackObserver)
}

type instrumentedAIClient struct {
	inner    aiplatform.Client
	recorder *Recorder
}

// WrapAIClient decorates AI client with latency/error metrics.
func WrapAIClient(inner aiplatform.Client, recorder *Recorder) aiplatform.Client {
	if inner == nil || recorder == nil {
		return inner
	}
	if observable, ok := inner.(fallbackObservable); ok {
		observable.SetFallbackObserver(recorder)
	}
	return &instrumentedAIClient{
		inner:    inner,
		recorder: recorder,
	}
}

// SetFallbackObserver delegates fallback observer configuration to wrapped client if supported.
func (c *instrumentedAIClient) SetFallbackObserver(observer aiplatform.FallbackObserver) {
	if c == nil || c.inner == nil {
		return
	}
	if observable, ok := c.inner.(fallbackObservable); ok {
		observable.SetFallbackObserver(observer)
	}
}

func (c *instrumentedAIClient) Generate(ctx context.Context, req aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
	if c == nil || c.inner == nil {
		return aiplatform.GenerateResponse{}, &aiplatform.ProviderError{Kind: aiplatform.ErrorKindUnavailable, Message: "ai client is not ready"}
	}

	started := time.Now()
	resp, err := c.inner.Generate(ctx, req)
	duration := time.Since(started)

	provider := strings.TrimSpace(string(resp.Provider))
	model := strings.TrimSpace(resp.Model)
	if provider == "" {
		provider = "unknown"
	}
	if model == "" {
		model = "unknown"
	}

	c.recorder.RecordAIRequest(provider, model, duration, err)
	return resp, err
}
