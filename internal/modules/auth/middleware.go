package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const principalContextKey = "recova.auth.principal"

// RequireAuth validates bearer access token and injects authenticated principal into request context.
func RequireAuth(service *Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		if service == nil {
			return errs.New(errs.CodeInternalError, "Layanan autentikasi belum siap", nil, nil)
		}

		token, err := ExtractBearerToken(c.Get("Authorization"))
		if err != nil {
			return err
		}

		principal, err := service.AuthenticateAccessToken(token)
		if err != nil {
			return err
		}

		c.Locals(principalContextKey, principal)
		return c.Next()
	}
}

// PrincipalFromContext retrieves authenticated principal from middleware context.
func PrincipalFromContext(c fiber.Ctx) (AuthPrincipal, bool) {
	value := c.Locals(principalContextKey)
	principal, ok := value.(AuthPrincipal)
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return AuthPrincipal{}, false
	}
	return principal, true
}

// ExtractBearerToken parses bearer token from Authorization header.
func ExtractBearerToken(header string) (string, error) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "", errs.New(errs.CodeUnauthenticated, "Token autentikasi wajib diisi", nil, nil)
	}

	const prefix = "bearer "
	if len(trimmed) < len(prefix) || strings.ToLower(trimmed[:len(prefix)]) != prefix {
		return "", errs.New(errs.CodeUnauthenticated, "Format token autentikasi tidak valid", nil, nil)
	}

	token := strings.TrimSpace(trimmed[len(prefix):])
	if token == "" {
		return "", errs.New(errs.CodeUnauthenticated, "Token autentikasi wajib diisi", nil, nil)
	}

	return token, nil
}
