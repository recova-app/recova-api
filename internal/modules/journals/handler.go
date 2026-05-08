package journals

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates journals HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs journals handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateJournal handles create-journal route.
func (h *Handler) CreateJournal(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req CreateJournalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CreateJournal(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Entri jurnal berhasil dibuat", payload, nil))
}

// ListJournals handles list-journals route.
func (h *Handler) ListJournals(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.ListJournals(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Jurnal berhasil diambil", payload, nil))
}
