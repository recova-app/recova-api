package users

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates users/onboarding HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs users handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns current authenticated user profile.
func (h *Handler) GetMe(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetCurrentUser(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Profil pengguna berhasil diambil", payload, nil))
}

// UpdateSettings updates current-user settings payload.
func (h *Handler) UpdateSettings(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req SettingsUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.UpdateSettings(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Pengaturan pengguna berhasil diperbarui", payload, nil))
}

// CompleteOnboarding stores onboarding payload for current authenticated user.
func (h *Handler) CompleteOnboarding(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req OnboardingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CompleteOnboarding(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Data onboarding berhasil disimpan", payload, nil))
}

// ResetUserDataForTesting resets development-only user generated data.
func (h *Handler) ResetUserDataForTesting(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	if err := h.service.ResetUserDataForTesting(c.Context(), principal.UserID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Data pengguna berhasil direset", fiber.Map{"reset": true}, nil))
}
