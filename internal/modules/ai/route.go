package ai

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers AI module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service, aiLimiter fiber.Handler) {
	handler := NewHandler(service)
	authGuard := authmodule.RequireAuth(authService)

	if aiLimiter == nil {
		router.Post("/ask-coach", authGuard, handler.AskCoach)
		router.Get("/chat-history", authGuard, handler.GetChatHistory)
		router.Get("/summary", authGuard, handler.GetSummary)
		router.Post("/onboarding-analysis", authGuard, handler.OnboardingAnalysis)
		router.Post("/relapse-solution", authGuard, handler.RelapseSolution)
		router.Get("/persona-preferences", authGuard, handler.GetPersonaPreference)
		router.Put("/persona-preferences", authGuard, handler.UpdatePersonaPreference)
		return
	}

	router.Post("/ask-coach", authGuard, aiLimiter, handler.AskCoach)
	router.Get("/chat-history", authGuard, aiLimiter, handler.GetChatHistory)
	router.Get("/summary", authGuard, aiLimiter, handler.GetSummary)
	router.Post("/onboarding-analysis", authGuard, aiLimiter, handler.OnboardingAnalysis)
	router.Post("/relapse-solution", authGuard, aiLimiter, handler.RelapseSolution)
	router.Get("/persona-preferences", authGuard, aiLimiter, handler.GetPersonaPreference)
	router.Put("/persona-preferences", authGuard, aiLimiter, handler.UpdatePersonaPreference)
}
