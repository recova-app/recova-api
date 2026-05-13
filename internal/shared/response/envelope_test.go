// Package response tests API envelope contract helpers.
package response

import "testing"

// TestSuccess_UsesContractShape ensures success helper fills standard shape.
func TestSuccess_UsesContractShape(t *testing.T) {
	data := map[string]any{"id": "user-1"}
	meta := map[string]any{"pagination": nil}

	got := Success("Data berhasil diambil", data, meta)
	if !got.Success {
		t.Fatal("expected success=true")
	}

	if got.Message != "Data berhasil diambil" {
		t.Fatalf("unexpected message: %s", got.Message)
	}

	if got.Data == nil {
		t.Fatal("expected data not nil")
	}

	if got.Meta == nil {
		t.Fatal("expected meta not nil")
	}
}

// TestSuccess_EmptyMessage_UsesDefault ensures helper keeps fallback message stable.
func TestSuccess_EmptyMessage_UsesDefault(t *testing.T) {
	got := Success("", map[string]any{}, nil)
	if got.Message != defaultSuccessMessage {
		t.Fatalf("unexpected fallback message: %s", got.Message)
	}
}

// TestError_UsesContractShape ensures error helper fills standard shape.
func TestError_UsesContractShape(t *testing.T) {
	details := []map[string]string{{"path": "email", "message": "format tidak valid"}}
	got := Error("Validasi data gagal", "VALIDATION_ERROR", details, "req_123")

	if got.Success {
		t.Fatal("expected success=false")
	}

	if got.Data != nil {
		t.Fatalf("expected data nil, got: %#v", got.Data)
	}

	if got.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("unexpected code: %s", got.Error.Code)
	}

	if got.Error.RequestID != "req_123" {
		t.Fatalf("unexpected request id: %s", got.Error.RequestID)
	}
}

// TestError_EmptyMessage_UsesDefault ensures default safe message is applied.
func TestError_EmptyMessage_UsesDefault(t *testing.T) {
	got := Error(" ", "INTERNAL_ERROR", nil, "")
	if got.Message != defaultErrorMessage {
		t.Fatalf("unexpected default message: %s", got.Message)
	}
}
