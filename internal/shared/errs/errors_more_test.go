package errs

import (
	"errors"
	"testing"
)

func TestAppError_ErrorAndUnwrap(t *testing.T) {
	cause := errors.New("root")
	err := New(CodeBadRequest, "bad", nil, cause)

	if err.Error() != "root" {
		t.Fatalf("unexpected error string: %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected errors.Is matches cause")
	}
	if err.Unwrap() != cause {
		t.Fatal("expected unwrap returns cause")
	}
}

func TestHTTPStatusByCode_CoversDefaultBranches(t *testing.T) {
	if HTTPStatusByCode(CodeBadRequest) != 400 {
		t.Fatal("bad request status mismatch")
	}
	if HTTPStatusByCode(CodeInternalError) != 500 {
		t.Fatal("internal error status mismatch")
	}
	if HTTPStatusByCode(Code("UNKNOWN")) != 500 {
		t.Fatal("unknown code should map to 500")
	}
}

func TestCodeFromHTTPStatus_CoversBranches(t *testing.T) {
	if codeFromHTTPStatus(401) != CodeUnauthenticated {
		t.Fatal("401 mapping mismatch")
	}
	if codeFromHTTPStatus(404) != CodeNotFound {
		t.Fatal("404 mapping mismatch")
	}
	if codeFromHTTPStatus(503) != CodeServiceUnavailable {
		t.Fatal("503 mapping should be non-empty")
	}
	if codeFromHTTPStatus(123) == "" {
		t.Fatal("unexpected empty code for 123")
	}
}
