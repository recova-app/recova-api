package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates auth HTTP requests to service operations.
type Handler struct {
	service *Service
}

// NewHandler constructs auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GoogleLogin handles Google login and session issuance.
func (h *Handler) GoogleLogin(c fiber.Ctx) error {
	var req GoogleLoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	result, err := h.service.LoginWithGoogle(c.Context(), req)
	if err != nil {
		return err
	}

	payload, err := h.service.BuildAuthResponseData(c.Context(), result.UserID, result.Session)
	if err != nil {
		return err
	}

	c.Cookie(h.service.RefreshCookie(result.RefreshToken))
	return c.Status(fiber.StatusOK).JSON(response.Success("Login berhasil", payload, nil))
}

// Register handles manual account registration and session issuance.
func (h *Handler) Register(c fiber.Ctx) error {
	var req ManualRegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	result, err := h.service.RegisterManual(c.Context(), req)
	if err != nil {
		return err
	}

	payload, err := h.service.BuildAuthResponseData(c.Context(), result.UserID, result.Session)
	if err != nil {
		return err
	}

	c.Cookie(h.service.RefreshCookie(result.RefreshToken))
	return c.Status(fiber.StatusCreated).JSON(response.Success("Registrasi berhasil", payload, nil))
}

// Login handles manual login by email or username.
func (h *Handler) Login(c fiber.Ctx) error {
	var req ManualLoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	result, err := h.service.LoginManual(c.Context(), req)
	if err != nil {
		return err
	}

	payload, err := h.service.BuildAuthResponseData(c.Context(), result.UserID, result.Session)
	if err != nil {
		return err
	}

	c.Cookie(h.service.RefreshCookie(result.RefreshToken))
	return c.Status(fiber.StatusOK).JSON(response.Success("Login berhasil", payload, nil))
}

// Refresh rotates refresh token cookie and issues new access token.
func (h *Handler) Refresh(c fiber.Ctx) error {
	refreshToken := h.service.RefreshCookieValue(c)
	result, err := h.service.RefreshSession(c.Context(), refreshToken)
	if err != nil {
		return err
	}

	payload, err := h.service.BuildAuthResponseData(c.Context(), result.UserID, result.Session)
	if err != nil {
		return err
	}

	c.Cookie(h.service.RefreshCookie(result.RefreshToken))
	return c.Status(fiber.StatusOK).JSON(response.Success("Sesi berhasil diperbarui", payload, nil))
}

// Logout revokes refresh session state and clears cookie.
func (h *Handler) Logout(c fiber.Ctx) error {
	refreshToken := h.service.RefreshCookieValue(c)
	if err := h.service.Logout(c.Context(), refreshToken); err != nil {
		return err
	}

	c.Cookie(h.service.ExpiredRefreshCookie())
	return c.Status(fiber.StatusOK).JSON(response.Success("Logout berhasil", fiber.Map{"logged_out": true}, nil))
}
