package ai

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates AI HTTP routes into service operations.
type Handler struct {
	service *Service
}

// NewHandler constructs AI handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// AskCoach handles ask-coach endpoint.
func (h *Handler) AskCoach(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req AskCoachRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.AskCoach(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Respon AI Coach berhasil dibuat", payload, nil))
}

// GetChatHistory handles chat-history endpoint.
func (h *Handler) GetChatHistory(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query ChatHistoryQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.GetChatHistory(c.Context(), principal.UserID, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Riwayat chat AI berhasil diambil", payload, nil))
}

// GetSummary handles summary endpoint.
func (h *Handler) GetSummary(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	payload, err := h.service.GetSummary(c.Context(), principal.UserID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Ringkasan pengguna berhasil diambil", payload, nil))
}

// OnboardingAnalysis handles onboarding-analysis endpoint.
func (h *Handler) OnboardingAnalysis(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req OnboardingAnalysisRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.AnalyzeOnboarding(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Analisis onboarding berhasil dibuat", payload, nil))
}
