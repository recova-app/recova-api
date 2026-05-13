package observability

// AIPersonaTelemetry bridges AI module persona telemetry into metrics recorder.
type AIPersonaTelemetry struct {
	recorder *Recorder
}

// NewAIPersonaTelemetry constructs persona telemetry recorder adapter.
func NewAIPersonaTelemetry(recorder *Recorder) *AIPersonaTelemetry {
	if recorder == nil {
		return nil
	}
	return &AIPersonaTelemetry{recorder: recorder}
}

// RecordPersonaUsage records persona usage events without prompt/response content.
func (t *AIPersonaTelemetry) RecordPersonaUsage(action string, persona string, err error) {
	if t == nil || t.recorder == nil {
		return
	}
	t.recorder.RecordAIPersonaUsage(action, persona, err)
}
