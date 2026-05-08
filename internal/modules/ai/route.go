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
		return
	}

	router.Post("/ask-coach", authGuard, aiLimiter, handler.AskCoach)
	router.Get("/chat-history", authGuard, aiLimiter, handler.GetChatHistory)
	router.Get("/summary", authGuard, aiLimiter, handler.GetSummary)
	router.Post("/onboarding-analysis", authGuard, aiLimiter, handler.OnboardingAnalysis)
}
