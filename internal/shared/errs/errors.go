// Package errs defines shared domain and platform error contracts.
package errs

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Code is a stable machine-readable API error code.
type Code string

const (
	// CodeBadRequest indicates malformed request input.
	CodeBadRequest Code = "BAD_REQUEST"
	// CodeValidationError indicates semantic input validation failure.
	CodeValidationError Code = "VALIDATION_ERROR"
	// CodeUnauthenticated indicates missing/invalid authentication.
	CodeUnauthenticated Code = "UNAUTHENTICATED"
	// CodeForbidden indicates access denied to authenticated user.
	CodeForbidden Code = "FORBIDDEN"
	// CodeNotFound indicates requested resource is not found.
	CodeNotFound Code = "NOT_FOUND"
	// CodeConflict indicates conflict with current resource state.
	CodeConflict Code = "CONFLICT"
	// CodeRateLimited indicates request rejected by rate limiter.
	CodeRateLimited Code = "RATE_LIMITED"
	// CodeDownstreamError indicates dependency/downstream failure.
	CodeDownstreamError Code = "DOWNSTREAM_ERROR"
	// CodeServiceUnavailable indicates service dependency unavailable.
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	// CodeInternalError indicates uncategorized internal failure.
	CodeInternalError Code = "INTERNAL_ERROR"
)

// MappedError is a safe API-facing error representation.
type MappedError struct {
	Status  int
	Code    Code
	Message string
	Details any
}

// AppError is an error type with explicit API mapping information.
type AppError struct {
	code    Code
	message string
	details any
	cause   error
}

// New builds a typed application error.
func New(code Code, message string, details any, cause error) *AppError {
	return &AppError{
		code:    code,
		message: strings.TrimSpace(message),
		details: details,
		cause:   cause,
	}
}

// Error returns the internal error string.
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.cause != nil {
		return e.cause.Error()
	}

	if e.message != "" {
		return e.message
	}

	return string(e.code)
}

// Unwrap returns wrapped cause for errors.Is/errors.As support.
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

// Code returns the stable error code.
func (e *AppError) Code() Code {
	if e == nil {
		return CodeInternalError
	}

	if e.code == "" {
		return CodeInternalError
	}

	return e.code
}

// Message returns a safe user-facing message in Bahasa Indonesia.
func (e *AppError) Message() string {
	if e == nil {
		return defaultMessageByCode(CodeInternalError)
	}

	if strings.TrimSpace(e.message) == "" {
		return defaultMessageByCode(e.Code())
	}

	return e.message
}

// Details returns optional safe error details.
func (e *AppError) Details() any {
	if e == nil {
		return nil
	}

	return e.details
}

// Map maps any runtime error to a safe API-facing mapped error.
func Map(err error) MappedError {
	if err == nil {
		return mappedFromCode(CodeInternalError, nil, "")
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return mappedFromCode(appErr.Code(), appErr.Details(), appErr.Message())
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code := codeFromHTTPStatus(fiberErr.Code)
		return mappedFromCode(code, nil, "")
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return mappedFromCode(CodeDownstreamError, nil, "")
	}

	if errors.Is(err, context.Canceled) {
		return mappedFromCode(CodeBadRequest, nil, "")
	}

	return mappedFromCode(CodeInternalError, nil, "")
}

// HTTPStatusByCode returns HTTP status associated with the given error code.
func HTTPStatusByCode(code Code) int {
	switch code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeValidationError:
		return http.StatusUnprocessableEntity
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeDownstreamError:
		return http.StatusBadGateway
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func mappedFromCode(code Code, details any, message string) MappedError {
	safeMessage := strings.TrimSpace(message)
	if safeMessage == "" {
		safeMessage = defaultMessageByCode(code)
	}

	return MappedError{
		Status:  HTTPStatusByCode(code),
		Code:    code,
		Message: safeMessage,
		Details: details,
	}
}

func codeFromHTTPStatus(status int) Code {
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnprocessableEntity:
		return CodeValidationError
	case http.StatusUnauthorized:
		return CodeUnauthenticated
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	case http.StatusTooManyRequests:
		return CodeRateLimited
	case http.StatusBadGateway:
		return CodeDownstreamError
	case http.StatusServiceUnavailable:
		return CodeServiceUnavailable
	default:
		if status >= 500 {
			return CodeInternalError
		}
		return CodeBadRequest
	}
}

func defaultMessageByCode(code Code) string {
	switch code {
	case CodeBadRequest:
		return "Permintaan tidak valid"
	case CodeValidationError:
		return "Validasi data gagal"
	case CodeUnauthenticated:
		return "Autentikasi dibutuhkan"
	case CodeForbidden:
		return "Akses ditolak"
	case CodeNotFound:
		return "Data tidak ditemukan"
	case CodeConflict:
		return "Terjadi konflik data"
	case CodeRateLimited:
		return "Terlalu banyak permintaan"
	case CodeDownstreamError:
		return "Layanan ketergantungan sedang bermasalah"
	case CodeServiceUnavailable:
		return "Layanan sedang tidak tersedia"
	default:
		return "Terjadi kesalahan pada server"
	}
}
