package auth

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

func TestRegisterCoreRoutes_GoogleLoginSuccess(t *testing.T) {
	reason := "pemulihan"
	service := NewService(&fakeAuthRepo{
		userByGoogle: models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester", UserWhy: &reason},
		userByID:     models.User{ID: "user-1", Email: "user@example.test", Nickname: "tester", UserWhy: &reason},
	}, &fakeGoogleVerifier{
		identity: GoogleIdentity{GoogleID: "google-1", Email: "user@example.test", DisplayName: "Tester"},
	}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/google", map[string]any{
		"token": "valid-google-token",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("expected refresh cookie set on login")
	}
}

func TestRegisterCoreRoutes_GoogleLoginValidationError(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/google", map[string]any{}, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterCoreRoutes_LogoutUnauthorizedWithoutBearer(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{parseAccessErr: errors.New("invalid")})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/logout", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func newAuthTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
