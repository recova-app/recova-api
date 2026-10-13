package routine

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers routine-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service) {
	handler := NewHandler(service)

	router.Use(authmodule.RequireAuth(authService))
	router.Post("/checkin", handler.DailyCheckIn)
	router.Post("/relapses", handler.CreateRelapse)
	router.Get("/statistics", handler.GetStatistics)
	router.Get("/statistics/activity-summary", handler.GetActivitySummary)
	router.Get("/relapses/statistics", handler.GetRelapseStatistics)
	router.Get("/relapses", handler.GetRelapses)
}
