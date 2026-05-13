package education

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates education HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs education handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListContents handles list-education route.
func (h *Handler) ListContents(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.ListContents(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Konten edukasi berhasil diambil", payload, nil))
}
