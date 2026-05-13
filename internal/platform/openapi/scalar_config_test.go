package openapi_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type scalarRoute struct {
	Type     string                 `json:"type"`
	Filepath string                 `json:"filepath"`
	Children map[string]scalarRoute `json:"children"`
}

type scalarConfig struct {
	Scalar     string `json:"scalar"`
	Navigation struct {
		Routes map[string]scalarRoute `json:"routes"`
	} `json:"navigation"`
}

func TestScalarConfig_MapsDocsAndOpenAPIArtifact(t *testing.T) {
	cfg := loadScalarConfig(t)

	if cfg.Scalar != "2.0.0" {
		t.Fatalf("expected scalar version 2.0.0, got: %s", cfg.Scalar)
	}

	apiRoute, ok := cfg.Navigation.Routes["/api"]
	if !ok {
		t.Fatal("expected /api route in scalar.config.json")
	}

	if apiRoute.Type != "openapi" {
		t.Fatalf("expected /api route type openapi, got: %s", apiRoute.Type)
	}

	if apiRoute.Filepath != "docs/generated/openapi.yaml" {
		t.Fatalf("expected /api filepath docs/generated/openapi.yaml, got: %s", apiRoute.Filepath)
	}

	filepaths := collectFilepaths(cfg.Navigation.Routes)
	for _, path := range filepaths {
		resolvedPath, err := resolveProjectPath(path)
		if err != nil {
			t.Fatalf("resolve route filepath: %s (%v)", path, err)
		}
		if _, err := os.Stat(resolvedPath); err != nil {
			t.Fatalf("expected route filepath to exist: %s (%v)", path, err)
		}
	}
}

func TestScalarConfig_OpenAPIArtifactIncludesRuntimeDocsRoutes(t *testing.T) {
	path, err := resolveProjectPath("docs/generated/openapi.yaml")
	if err != nil {
		t.Fatalf("resolve openapi artifact path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read openapi artifact: %v", err)
	}

	text := string(raw)
	if !strings.Contains(text, "/openapi.yaml:") {
		t.Fatalf("expected /openapi.yaml route in docs/generated/openapi.yaml")
	}
	if !strings.Contains(text, "/docs/api:") {
		t.Fatalf("expected /docs/api route in docs/generated/openapi.yaml")
	}
}

func loadScalarConfig(t testing.TB) scalarConfig {
	t.Helper()

	path, err := resolveProjectPath("scalar.config.json")
	if err != nil {
		t.Fatalf("resolve scalar.config.json path: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read scalar.config.json: %v", err)
	}

	var cfg scalarConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse scalar.config.json: %v", err)
	}

	return cfg
}

func collectFilepaths(routes map[string]scalarRoute) []string {
	paths := make([]string, 0, len(routes))
	for _, route := range routes {
		switch route.Type {
		case "page", "openapi":
			paths = append(paths, route.Filepath)
		case "group":
			paths = append(paths, collectFilepaths(route.Children)...)
		}
	}
	return paths
}

func resolveProjectPath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		_, err := os.Stat(cleanPath)
		return cleanPath, err
	}

	if _, err := os.Stat(cleanPath); err == nil {
		return cleanPath, nil
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := workingDir
	for {
		candidate := filepath.Join(current, cleanPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", errors.New("file not found from project root lookup")
}
