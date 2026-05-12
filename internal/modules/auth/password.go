package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(rawPassword string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

func verifyPassword(passwordHash string, rawPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(rawPassword)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}
