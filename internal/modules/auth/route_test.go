package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/recova-app/backend-v2/internal/platform/database/models"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
	"gorm.io/gorm"
)

func TestRegisterCoreRoutes_GoogleLoginSuccess(t *testing.T) {
	reason := "recovery"
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

func TestRegisterCoreRoutes_ManualRegisterSuccess(t *testing.T) {
	reason := "recovery"
	service := NewService(&fakeAuthRepo{
		manualCreatedUser: models.User{
			ID:       "user-manual-1",
			Email:    "manual@example.test",
			Nickname: "manualuser",
		},
		userByID: models.User{
			ID:       "user-manual-1",
			Email:    "manual@example.test",
			Nickname: "manualuser",
			UserWhy:  &reason,
		},
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":            "manual@example.test",
		"username":         "manualuser",
		"password":         "password123",
		"confirm_password": "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusCreated)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("expected refresh cookie set on register")
	}
}

func TestRegisterCoreRoutes_ManualRegisterPasswordMismatch(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":            "manual@example.test",
		"username":         "manualuser",
		"password":         "password123",
		"confirm_password": "password456",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterCoreRoutes_ManualRegisterWeakPassword(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":            "manual@example.test",
		"username":         "manualuser",
		"password":         "short",
		"confirm_password": "short",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterCoreRoutes_ManualRegisterDuplicateEmail(t *testing.T) {
	service := NewService(&fakeAuthRepo{
		createManualErr: &pgconn.PgError{Code: "23505", ConstraintName: "uq_users_email"},
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":            "manual@example.test",
		"username":         "manualuser",
		"password":         "password123",
		"confirm_password": "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusConflict)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "CONFLICT")
}

func TestRegisterCoreRoutes_ManualRegisterDuplicateUsername(t *testing.T) {
	service := NewService(&fakeAuthRepo{
		createManualErr: &pgconn.PgError{Code: "23505", ConstraintName: "uq_users_username"},
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/register", map[string]any{
		"email":            "manual@example.test",
		"username":         "manualuser",
		"password":         "password123",
		"confirm_password": "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusConflict)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "CONFLICT")
}

func TestRegisterCoreRoutes_ManualLoginValidationError(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/login", map[string]any{
		"password": "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "VALIDATION_ERROR")
}

func TestRegisterCoreRoutes_ManualLoginNotFound(t *testing.T) {
	service := NewService(&fakeAuthRepo{findManualErr: gorm.ErrRecordNotFound}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/login", map[string]any{
		"identifier": "missing@example.test",
		"password":   "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterCoreRoutes_ManualLoginWrongPassword(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	username := "manualuser"
	service := NewService(&fakeAuthRepo{
		manualUserByIdentifier: models.User{
			ID:           "user-manual-1",
			Email:        "manual@example.test",
			Username:     &username,
			PasswordHash: &hash,
			Nickname:     "manualuser",
		},
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/login", map[string]any{
		"identifier": "manual@example.test",
		"password":   "wrong-password",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterCoreRoutes_ManualLoginSuccess(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	reason := "recovery"
	username := "manualuser"
	user := models.User{
		ID:           "user-manual-1",
		Email:        "manual@example.test",
		Username:     &username,
		PasswordHash: &hash,
		Nickname:     "manualuser",
		UserWhy:      &reason,
	}

	service := NewService(&fakeAuthRepo{
		userByID:               user,
		manualUserByIdentifier: user,
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/login", map[string]any{
		"identifier": "manual@example.test",
		"password":   "password123",
	}, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("expected refresh cookie set on login")
	}
}

func TestRegisterCoreRoutes_RefreshSuccess_RotatesCookieAndTokenState(t *testing.T) {
	reason := "recovery"
	repo := &fakeAuthRepo{
		userByID: models.User{
			ID:       "user-1",
			Email:    "user@example.test",
			Nickname: "tester",
			UserWhy:  &reason,
		},
		storedRefresh: models.AuthRefreshToken{
			ID:        "refresh-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	service := NewService(repo, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/refresh", nil, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("expected refresh cookie set on refresh")
	}
	if repo.rotatedFromTokenID != "refresh-1" {
		t.Fatalf("expected refresh rotation called with old token id, got=%q", repo.rotatedFromTokenID)
	}
}

func TestRegisterCoreRoutes_RefreshExpired_ReturnsUnauthorized(t *testing.T) {
	reason := "recovery"
	service := NewService(&fakeAuthRepo{
		userByID: models.User{
			ID:       "user-1",
			Email:    "user@example.test",
			Nickname: "tester",
			UserWhy:  &reason,
		},
		storedRefresh: models.AuthRefreshToken{
			ID:        "refresh-1",
			UserID:    "user-1",
			TokenHash: "hash-refresh",
			ExpiresAt: time.Now().Add(-1 * time.Minute),
		},
	}, &fakeGoogleVerifier{}, &fakeTokenProvider{
		parseRefreshClaims: SessionClaims{RegisteredClaims: jwtRegisteredClaims("user-1", "refresh-jti", time.Now().Add(-1*time.Minute))},
	})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/refresh", nil, nil)

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterCoreRoutes_LogoutUnauthorizedWithoutBearer(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{parseAccessErr: errors.New("invalid")})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/logout", nil, nil)
	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusUnauthorized)
	httpharness.RequireErrorEnvelope(t, resp.JSON, "UNAUTHENTICATED")
}

func TestRegisterCoreRoutes_LogoutSuccess_ClearsCookie(t *testing.T) {
	service := NewService(&fakeAuthRepo{}, &fakeGoogleVerifier{}, &fakeTokenProvider{})

	app := newAuthTestApp()
	RegisterCoreRoutes(app.Group("/api/v1/auth"), service)

	resp := httpharness.JSONRequest(t, app, fiber.MethodPost, "/api/v1/auth/logout", nil, map[string]string{
		"Authorization": "Bearer access-token",
	})

	httpharness.RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	httpharness.RequireSuccessEnvelope(t, resp.JSON)
	if resp.Header.Get("Set-Cookie") == "" {
		t.Fatal("expected expired refresh cookie set on logout")
	}
}

func newAuthTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			mapped := errs.Map(err)
			return c.Status(mapped.Status).JSON(response.Error(mapped.Message, string(mapped.Code), mapped.Details, ""))
		},
	})
}
