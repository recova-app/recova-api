package achievements

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers achievements-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service) {
	handler := NewHandler(service)

	router.Get("/catalog", authmodule.RequireAuth(authService), handler.GetCatalog)
	router.Get("/progress", authmodule.RequireAuth(authService), handler.GetProgress)
	router.Get("/unlocked", authmodule.RequireAuth(authService), handler.GetUnlocked)
}
