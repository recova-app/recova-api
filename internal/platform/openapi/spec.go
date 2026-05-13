// Package openapi provides OpenAPI loading, validation, and route-contract helpers.
package openapi

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// RouteKey represents one HTTP method and normalized path tuple.
type RouteKey struct {
	Method string
	Path   string
}

// String returns stable string representation for diagnostics.
func (r RouteKey) String() string {
	return r.Method + " " + r.Path
}

// LoadAndValidate loads OpenAPI document from path and validates it.
func LoadAndValidate(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load openapi file: %w", err)
	}

	validateOpts := []openapi3.ValidationOption{}
	if doc.IsOpenAPI31OrLater() {
		validateOpts = append(validateOpts, openapi3.IsOpenAPI31OrLater())
	}
	if err := doc.Validate(context.Background(), validateOpts...); err != nil {
		return nil, fmt.Errorf("validate openapi file: %w", err)
	}

	return doc, nil
}

// ReadSpecBytes reads OpenAPI source file as bytes.
func ReadSpecBytes(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}
	return content, nil
}

// SpecRouteSet returns normalized method-path set from OpenAPI document.
func SpecRouteSet(doc *openapi3.T) map[RouteKey]struct{} {
	routes := map[RouteKey]struct{}{}
	if doc == nil || doc.Paths == nil {
		return routes
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Value(path)
		if pathItem == nil {
			continue
		}

		for method := range pathItem.Operations() {
			key := RouteKey{
				Method: strings.ToUpper(strings.TrimSpace(method)),
				Path:   NormalizeOpenAPIPath(path),
			}
			if key.Method == "" || key.Path == "" {
				continue
			}
			routes[key] = struct{}{}
		}
	}

	return routes
}

// SortedRouteKeys returns route keys sorted lexicographically by method then path.
func SortedRouteKeys(routeSet map[RouteKey]struct{}) []RouteKey {
	keys := make([]RouteKey, 0, len(routeSet))
	for key := range routeSet {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Method == keys[j].Method {
			return keys[i].Path < keys[j].Path
		}
		return keys[i].Method < keys[j].Method
	})
	return keys
}
