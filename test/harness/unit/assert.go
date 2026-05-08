// Package unit provides minimal reusable assertions for unit-level tests.
package unit

import (
	"reflect"
	"testing"
)

// RequireNoError fails the test when err is not nil.
func RequireNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// RequireError fails the test when err is nil.
func RequireError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// RequireEqual fails the test when got and want are not deeply equal.
func RequireEqual(t testing.TB, want any, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("value mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}
