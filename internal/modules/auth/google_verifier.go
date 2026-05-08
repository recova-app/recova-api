package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"
)

var allowedGoogleIssuers = map[string]struct{}{
	"accounts.google.com":         {},
	"https://accounts.google.com": {},
}

// GoogleTokenVerifier verifies Google ID token and extracts identity claims.
type GoogleTokenVerifier interface {
	Verify(ctx context.Context, rawToken string, audience string) (GoogleIdentity, error)
}

// GoogleIDTokenVerifier validates token using Google public key material.
type GoogleIDTokenVerifier struct{}

// Verify validates Google ID token and returns safe identity claims.
func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, rawToken string, audience string) (GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, strings.TrimSpace(rawToken), strings.TrimSpace(audience))
	if err != nil {
		return GoogleIdentity{}, fmt.Errorf("google token invalid: %w", err)
	}

	if _, ok := allowedGoogleIssuers[strings.TrimSpace(payload.Issuer)]; !ok {
		return GoogleIdentity{}, fmt.Errorf("google issuer tidak valid")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	if strings.TrimSpace(payload.Subject) == "" {
		return GoogleIdentity{}, fmt.Errorf("google subject kosong")
	}
	if strings.TrimSpace(email) == "" {
		return GoogleIdentity{}, fmt.Errorf("email Google tidak tersedia")
	}

	return GoogleIdentity{
		GoogleID:    strings.TrimSpace(payload.Subject),
		Email:       strings.ToLower(strings.TrimSpace(email)),
		DisplayName: strings.TrimSpace(name),
	}, nil
}
