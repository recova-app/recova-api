package openapi

import (
	"regexp"
	"strings"
)

var (
	fiberNamedParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)\??`)
	multiSlashPattern      = regexp.MustCompile(`/+`)
)

// NormalizeOpenAPIPath normalizes OpenAPI path for comparisons.
func NormalizeOpenAPIPath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = multiSlashPattern.ReplaceAllString(p, "/")
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// FiberPathToOpenAPIPath converts Fiber route syntax to OpenAPI template syntax.
func FiberPathToOpenAPIPath(path string) string {
	normalized := NormalizeOpenAPIPath(path)
	if normalized == "" {
		return ""
	}

	converted := fiberNamedParamPattern.ReplaceAllString(normalized, `{$1}`)
	converted = strings.ReplaceAll(converted, "/*", "/{wildcard}")
	converted = strings.ReplaceAll(converted, "/+", "/{greedy}")
	return converted
}
