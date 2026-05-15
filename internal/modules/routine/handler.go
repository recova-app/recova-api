package routine

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates routine HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs routine handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// DailyCheckIn handles daily check-in creation.
func (h *Handler) DailyCheckIn(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req DailyCheckInRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CreateDailyCheckIn(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Check-in berhasil", payload, nil))
}

// CreateRelapse handles explicit relapse submission for current UTC day.
func (h *Handler) CreateRelapse(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req RelapseRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CreateRelapse(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Relapse berhasil dicatat", payload, nil))
}

// GetStatistics handles routine statistics retrieval.
func (h *Handler) GetStatistics(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetStatistics(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Statistik berhasil diambil", payload, nil))
}

// GetActivitySummary handles periodic activity summary retrieval.
func (h *Handler) GetActivitySummary(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query ActivitySummaryQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.GetActivitySummary(c.Context(), principal.UserID, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Ringkasan aktivitas berhasil diambil", payload, nil))
}

// GetRelapses handles relapse history retrieval.
func (h *Handler) GetRelapses(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetRelapses(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Riwayat relapse berhasil diambil", payload, nil))
}

// GetRelapseStatistics handles complete relapse statistics retrieval.
func (h *Handler) GetRelapseStatistics(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetRelapseStatistics(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Statistik relapse berhasil diambil", payload, nil))
}
