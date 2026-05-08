package users

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterUserRoutes registers users-module routes.
func RegisterUserRoutes(router fiber.Router, authService *authmodule.Service, usersService *Service) {
	handler := NewHandler(usersService)

	router.Get("/me", authmodule.RequireAuth(authService), handler.GetMe)
	router.Put("/settings", authmodule.RequireAuth(authService), handler.UpdateSettings)
	router.Delete("/me/reset-data", authmodule.RequireAuth(authService), handler.ResetUserDataForTesting)
}

// RegisterOnboardingRoute registers onboarding route under auth prefix.
func RegisterOnboardingRoute(authRouter fiber.Router, authService *authmodule.Service, usersService *Service) {
	handler := NewHandler(usersService)
	authRouter.Post("/onboarding", authmodule.RequireAuth(authService), handler.CompleteOnboarding)
}
