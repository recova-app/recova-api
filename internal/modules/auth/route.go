package auth

import "github.com/gofiber/fiber/v3"

// RegisterCoreRoutes registers auth routes owned by auth module.
func RegisterCoreRoutes(router fiber.Router, service *Service) {
	handler := NewHandler(service)

	router.Post("/google", handler.GoogleLogin)
	router.Post("/refresh", handler.Refresh)
	router.Post("/logout", RequireAuth(service), handler.Logout)
}
