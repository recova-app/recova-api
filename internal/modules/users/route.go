package users

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterUserRoutes registers users-module routes.
func RegisterUserRoutes(router fiber.Router, authService *authmodule.Service, usersService *Service) {
	handler := NewHandler(usersService)

	router.Use(authmodule.RequireAuth(authService))
	router.Get("/me", handler.GetMe)
	router.Put("/settings", handler.UpdateSettings)
	router.Delete("/me/reset-data", handler.ResetUserDataForTesting)
}

// RegisterOnboardingRoute registers onboarding route under auth prefix.
func RegisterOnboardingRoute(authRouter fiber.Router, authService *authmodule.Service, usersService *Service) {
	handler := NewHandler(usersService)
	authRouter.Post("/onboarding", authmodule.RequireAuth(authService), handler.CompleteOnboarding)
}
