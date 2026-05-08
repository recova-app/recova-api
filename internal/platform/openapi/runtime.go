package openapi

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

var allowedRuntimeMethods = map[string]struct{}{
	fiber.MethodGet:     {},
	fiber.MethodPost:    {},
	fiber.MethodPut:     {},
	fiber.MethodPatch:   {},
	fiber.MethodDelete:  {},
	fiber.MethodOptions: {},
}

// RuntimeRouteSet converts Fiber runtime routes into normalized OpenAPI-compatible keys.
func RuntimeRouteSet(routes []fiber.Route) map[RouteKey]struct{} {
	result := map[RouteKey]struct{}{}

	for _, route := range routes {
		method := strings.ToUpper(strings.TrimSpace(route.Method))
		if _, ok := allowedRuntimeMethods[method]; !ok {
			continue
		}

		path := FiberPathToOpenAPIPath(route.Path)
		if path == "" {
			continue
		}

		result[RouteKey{Method: method, Path: path}] = struct{}{}
	}

	return result
}
