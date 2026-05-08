package community

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
	"github.com/recova-app/backend-v2/internal/shared/response"
)

// Handler translates community HTTP routes into service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs community handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListPosts handles community feed listing.
func (h *Handler) ListPosts(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}
	if strings.TrimSpace(principal.UserID) == "" {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var query ListPostsQuery
	if err := c.Bind().Query(&query); err != nil {
		return err
	}

	payload, err := h.service.ListPosts(c.Context(), query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Postingan komunitas berhasil diambil", payload, nil))
}

// CreatePost handles community post creation.
func (h *Handler) CreatePost(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	var req CreatePostRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CreatePost(c.Context(), principal.UserID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Postingan komunitas berhasil dibuat", payload, nil))
}

// CreateComment handles comment creation on selected post.
func (h *Handler) CreateComment(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	postID := strings.TrimSpace(c.Params("postId"))
	if postID == "" {
		return errs.New(errs.CodeValidationError, "ID postingan wajib diisi", []map[string]string{{
			"field": "postId", "message": "ID postingan wajib diisi",
		}}, nil)
	}

	var req CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return err
	}

	payload, err := h.service.CreateComment(c.Context(), principal.UserID, postID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Komentar berhasil dibuat", payload, nil))
}

// ToggleLike handles like/unlike toggle on selected post.
func (h *Handler) ToggleLike(c fiber.Ctx) error {
	principal, ok := authmodule.PrincipalFromContext(c)
	if !ok {
		return errs.New(errs.CodeUnauthenticated, "Autentikasi dibutuhkan", nil, nil)
	}

	postID := strings.TrimSpace(c.Params("postId"))
	if postID == "" {
		return errs.New(errs.CodeValidationError, "ID postingan wajib diisi", []map[string]string{{
			"field": "postId", "message": "ID postingan wajib diisi",
		}}, nil)
	}

	payload, err := h.service.ToggleLike(c.Context(), principal.UserID, postID)
	if err != nil {
		return err
	}

	message := "Suka pada postingan dibatalkan"
	if payload.IsLiked {
		message = "Postingan berhasil disukai"
	}

	return c.Status(fiber.StatusOK).JSON(response.Success(message, payload, nil))
}
