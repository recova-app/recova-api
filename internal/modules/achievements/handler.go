package achievements

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates achievements HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs achievements handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetCatalog handles achievement catalog retrieval.
func (h *Handler) GetCatalog(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query CategoryQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.GetCatalog(c.Context(), principal.UserID, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Katalog achievement berhasil diambil", payload, nil))
}

// GetProgress handles achievement progress retrieval.
func (h *Handler) GetProgress(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query CategoryQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.GetProgress(c.Context(), principal.UserID, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Progres achievement berhasil diambil", payload, nil))
}

// GetUnlocked handles unlocked achievement retrieval.
func (h *Handler) GetUnlocked(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query CategoryQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.GetUnlocked(c.Context(), principal.UserID, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Daftar achievement terbuka berhasil diambil", payload, nil))
}
