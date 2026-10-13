package community

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers community-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service, writeLimiter fiber.Handler) {
	handler := NewHandler(service)

	router.Use(authmodule.RequireAuth(authService))
	router.Get("/", handler.ListPosts)
	router.Get("/:post_id/comments", handler.ListCommentThread)
	if writeLimiter == nil {
		router.Post("/", handler.CreatePost)
		router.Post("/:post_id/comments", handler.CreateComment)
		router.Post("/:post_id/comments/:comment_id/replies", handler.CreateReply)
		router.Post("/:post_id/like", handler.ToggleLike)
		return
	}

	router.Post("/", writeLimiter, handler.CreatePost)
	router.Post("/:post_id/comments", writeLimiter, handler.CreateComment)
	router.Post("/:post_id/comments/:comment_id/replies", writeLimiter, handler.CreateReply)
	router.Post("/:post_id/like", writeLimiter, handler.ToggleLike)
}
