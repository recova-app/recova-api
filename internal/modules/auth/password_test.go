package auth

import "testing"

func TestHashPasswordAndVerifyPassword(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if err := verifyPassword(hash, "password123"); err != nil {
		t.Fatalf("verify correct password: %v", err)
	}
	if err := verifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected verify error for mismatched password")
	}
}
