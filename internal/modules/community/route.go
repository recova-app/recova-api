package community

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers community-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service, writeLimiter fiber.Handler) {
	handler := NewHandler(service)

	router.Get("/", authmodule.RequireAuth(authService), handler.ListPosts)
	router.Get("/:post_id/comments", authmodule.RequireAuth(authService), handler.ListCommentThread)
	if writeLimiter == nil {
		router.Post("/", authmodule.RequireAuth(authService), handler.CreatePost)
		router.Post("/:post_id/comments", authmodule.RequireAuth(authService), handler.CreateComment)
		router.Post("/:post_id/comments/:comment_id/replies", authmodule.RequireAuth(authService), handler.CreateReply)
		router.Post("/:post_id/like", authmodule.RequireAuth(authService), handler.ToggleLike)
		return
	}

	router.Post("/", authmodule.RequireAuth(authService), writeLimiter, handler.CreatePost)
	router.Post("/:post_id/comments", authmodule.RequireAuth(authService), writeLimiter, handler.CreateComment)
	router.Post("/:post_id/comments/:comment_id/replies", authmodule.RequireAuth(authService), writeLimiter, handler.CreateReply)
	router.Post("/:post_id/like", authmodule.RequireAuth(authService), writeLimiter, handler.ToggleLike)
}
