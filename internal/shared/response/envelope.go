// Package response defines shared response contracts and helper formatters.
package response

import "strings"

const (
	defaultSuccessMessage = "Permintaan berhasil diproses"
	defaultErrorMessage   = "Permintaan gagal diproses"
)

// SuccessEnvelope defines API success response shape.
type SuccessEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta"`
}

// ErrorEnvelope defines API error response shape.
type ErrorEnvelope struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    any          `json:"data"`
	Error   ErrorPayload `json:"error"`
}

// ErrorPayload defines machine-readable API error details.
type ErrorPayload struct {
	Code      string `json:"code"`
	Details   any    `json:"details"`
	RequestID string `json:"request_id"`
}

// Success builds a standard success envelope.
func Success(message string, data any, meta any) SuccessEnvelope {
	safeMessage := strings.TrimSpace(message)
	if safeMessage == "" {
		safeMessage = defaultSuccessMessage
	}

	return SuccessEnvelope{
		Success: true,
		Message: safeMessage,
		Data:    data,
		Meta:    meta,
	}
}

// Error builds a standard error envelope.
func Error(message string, code string, details any, requestID string) ErrorEnvelope {
	safeMessage := strings.TrimSpace(message)
	if safeMessage == "" {
		safeMessage = defaultErrorMessage
	}

	return ErrorEnvelope{
		Success: false,
		Message: safeMessage,
		Data:    nil,
		Error: ErrorPayload{
			Code:      strings.TrimSpace(code),
			Details:   details,
			RequestID: strings.TrimSpace(requestID),
		},
	}
}
