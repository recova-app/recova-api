package httpharness

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestJSONRequest_AndEnvelopeHelpers(t *testing.T) {
	app := fiber.New()
	app.Get("/ok", func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "ok",
			"data":    fiber.Map{"status": "ok"},
		})
	})

	resp := JSONRequest(t, app, fiber.MethodGet, "/ok", nil, nil)
	RequireStatus(t, resp.StatusCode, fiber.StatusOK)
	RequireSuccessEnvelope(t, resp.JSON)
}

func TestMustPath_AcceptsValidPath(t *testing.T) {
	path := MustPath(t, "/api/v1")
	if path != "/api/v1" {
		t.Fatalf("unexpected normalized path: %s", path)
	}
}
