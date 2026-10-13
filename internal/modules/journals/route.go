package journals

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers journals-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service) {
	handler := NewHandler(service)

	router.Use(authmodule.RequireAuth(authService))
	router.Get("/", handler.ListJournals)
	router.Post("/", handler.CreateJournal)
}
