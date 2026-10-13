package achievements

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers achievements-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service) {
	handler := NewHandler(service)

	router.Use(authmodule.RequireAuth(authService))
	router.Get("/catalog", handler.GetCatalog)
	router.Get("/progress", handler.GetProgress)
	router.Get("/unlocked", handler.GetUnlocked)
}
