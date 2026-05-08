package routine

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers routine-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service) {
	handler := NewHandler(service)

	router.Post("/checkin", authmodule.RequireAuth(authService), handler.DailyCheckIn)
	router.Get("/statistics", authmodule.RequireAuth(authService), handler.GetStatistics)
	router.Get("/relapses", authmodule.RequireAuth(authService), handler.GetRelapses)
}
