// Package errs tests error taxonomy mapping behavior.
package errs

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// TestMap_AppError_ReturnsMappedError ensures typed app errors map predictably.
func TestMap_AppError_ReturnsMappedError(t *testing.T) {
	details := []map[string]string{{"path": "email", "message": "email tidak valid"}}
	err := New(CodeValidationError, "Input tidak valid", details, nil)

	mapped := Map(err)
	if mapped.Status != 422 {
		t.Fatalf("unexpected status: %d", mapped.Status)
	}

	if mapped.Code != CodeValidationError {
		t.Fatalf("unexpected code: %s", mapped.Code)
	}

	if mapped.Message != "Input tidak valid" {
		t.Fatalf("unexpected message: %s", mapped.Message)
	}

	if mapped.Details == nil {
		t.Fatal("expected details not nil")
	}
}

// TestMap_FiberError_MapsStatus ensures fiber errors map via status code taxonomy.
func TestMap_FiberError_MapsStatus(t *testing.T) {
	err := fiber.NewError(fiber.StatusNotFound, "raw internal text")

	mapped := Map(err)
	if mapped.Status != 404 {
		t.Fatalf("unexpected status: %d", mapped.Status)
	}

	if mapped.Code != CodeNotFound {
		t.Fatalf("unexpected code: %s", mapped.Code)
	}

	if mapped.Message == "raw internal text" {
		t.Fatal("expected safe default message, got raw fiber error message")
	}
}

// TestMap_ContextDeadlineExceeded_ReturnsDownstreamError ensures timeout maps to downstream class.
func TestMap_ContextDeadlineExceeded_ReturnsDownstreamError(t *testing.T) {
	mapped := Map(context.DeadlineExceeded)
	if mapped.Code != CodeDownstreamError {
		t.Fatalf("unexpected code: %s", mapped.Code)
	}

	if mapped.Status != 502 {
		t.Fatalf("unexpected status: %d", mapped.Status)
	}
}

// TestMap_UnknownError_ReturnsInternalError ensures unknown errors stay safe.
func TestMap_UnknownError_ReturnsInternalError(t *testing.T) {
	mapped := Map(errors.New("db: password=super-secret"))
	if mapped.Code != CodeInternalError {
		t.Fatalf("unexpected code: %s", mapped.Code)
	}

	if mapped.Status != 500 {
		t.Fatalf("unexpected status: %d", mapped.Status)
	}

	if mapped.Message == "db: password=super-secret" {
		t.Fatal("unsafe raw message leaked")
	}
}
