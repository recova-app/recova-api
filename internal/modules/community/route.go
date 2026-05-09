package community

import (
	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
)

// RegisterRoutes registers community-module routes.
func RegisterRoutes(router fiber.Router, authService *authmodule.Service, service *Service, writeLimiter fiber.Handler) {
	handler := NewHandler(service)

	router.Get("/", authmodule.RequireAuth(authService), handler.ListPosts)
	router.Get("/:postId/comments", authmodule.RequireAuth(authService), handler.ListCommentThread)
	if writeLimiter == nil {
		router.Post("/", authmodule.RequireAuth(authService), handler.CreatePost)
		router.Post("/:postId/comments", authmodule.RequireAuth(authService), handler.CreateComment)
		router.Post("/:postId/comments/:commentId/replies", authmodule.RequireAuth(authService), handler.CreateReply)
		router.Post("/:postId/like", authmodule.RequireAuth(authService), handler.ToggleLike)
		return
	}

	router.Post("/", authmodule.RequireAuth(authService), writeLimiter, handler.CreatePost)
	router.Post("/:postId/comments", authmodule.RequireAuth(authService), writeLimiter, handler.CreateComment)
	router.Post("/:postId/comments/:commentId/replies", authmodule.RequireAuth(authService), writeLimiter, handler.CreateReply)
	router.Post("/:postId/like", authmodule.RequireAuth(authService), writeLimiter, handler.ToggleLike)
}
