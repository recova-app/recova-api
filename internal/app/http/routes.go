package http

import "github.com/gofiber/fiber/v3"

// Routes returns registered routes from Fiber runtime.
// When excludeMiddleware is true, middleware routes are filtered out.
func (s *Server) Routes(excludeMiddleware bool) []fiber.Route {
	if s == nil || s.app == nil {
		return nil
	}
	if excludeMiddleware {
		return s.app.GetRoutes(true)
	}
	return s.app.GetRoutes()
}
