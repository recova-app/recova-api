package ai

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers AI module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service, aiLimiter fiber.Handler) {
	handler := NewHandler(service)

	router.Use(authmodule.RequireAuth(authService))
	if aiLimiter != nil {
		router.Use(aiLimiter)
	}

	router.Post("/ask-coach", handler.AskCoach)
	router.Get("/chat-history", handler.GetChatHistory)
	router.Get("/summary", handler.GetSummary)
	router.Post("/onboarding-analysis", handler.OnboardingAnalysis)
	router.Post("/relapse-solution", handler.RelapseSolution)
	router.Get("/persona-preferences", handler.GetPersonaPreference)
	router.Put("/persona-preferences", handler.UpdatePersonaPreference)
}
