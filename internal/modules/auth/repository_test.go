package auth

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestIsRecordNotFound(t *testing.T) {
	if !IsRecordNotFound(gorm.ErrRecordNotFound) {
		t.Fatal("expected true for gorm.ErrRecordNotFound")
	}
	if IsRecordNotFound(errors.New("other error")) {
		t.Fatal("expected false for non record-not-found error")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	uniqueErr := &pgconn.PgError{Code: uniqueViolationCode}
	if !IsUniqueViolation(uniqueErr) {
		t.Fatal("expected true for unique violation error code")
	}
	if IsUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Fatal("expected false for non-unique violation code")
	}
	if IsUniqueViolation(errors.New("plain error")) {
		t.Fatal("expected false for non pg error")
	}
}

func TestUniqueViolationConstraint(t *testing.T) {
	err := &pgconn.PgError{Code: uniqueViolationCode, ConstraintName: "uq_users_username"}
	if got := UniqueViolationConstraint(err); got != "uq_users_username" {
		t.Fatalf("unexpected constraint: %q", got)
	}
	if got := UniqueViolationConstraint(errors.New("plain")); got != "" {
		t.Fatalf("expected empty constraint, got: %q", got)
	}
}

func TestFallbackNickname(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "empty email", email: "", want: "User"},
		{name: "short local part", email: "ab@example.test", want: "User"},
		{name: "valid local part", email: "tester@example.test", want: "tester"},
		{name: "trim and lowercase", email: "  TESTER@EXAMPLE.TEST ", want: "tester"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := fallbackNickname(tc.email); got != tc.want {
				t.Fatalf("unexpected fallback nickname: got=%q want=%q", got, tc.want)
			}
		})
	}
}
