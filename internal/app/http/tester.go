package http

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// Test executes one synthetic HTTP request against the in-memory Fiber app.
func (s *Server) Test(req *http.Request, cfg ...fiber.TestConfig) (*http.Response, error) {
	if s == nil || s.app == nil {
		return nil, fmt.Errorf("server is not initialized")
	}
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	return s.app.Test(req, cfg...)
}
