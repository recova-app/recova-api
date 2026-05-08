package observability

import (
	"context"
	"strings"
	"time"

	aiplatform "github.com/recova-app/backend-v2/internal/platform/ai"
)

type instrumentedAIClient struct {
	inner    aiplatform.Client
	recorder *Recorder
}

// WrapAIClient decorates AI client with latency/error metrics.
func WrapAIClient(inner aiplatform.Client, recorder *Recorder) aiplatform.Client {
	if inner == nil || recorder == nil {
		return inner
	}
	return &instrumentedAIClient{
		inner:    inner,
		recorder: recorder,
	}
}

func (c *instrumentedAIClient) Generate(ctx context.Context, req aiplatform.GenerateRequest) (aiplatform.GenerateResponse, error) {
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
