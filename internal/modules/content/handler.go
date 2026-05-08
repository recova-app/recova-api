package content

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates daily content HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs daily content handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetDailyContent handles daily content retrieval.
func (h *Handler) GetDailyContent(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetDailyContent(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Konten harian berhasil diambil", payload, nil))
}
