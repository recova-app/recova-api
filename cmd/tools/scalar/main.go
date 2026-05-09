package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configPath                 = "scalar.config.json"
	requiredScalarVersion      = "2.0.0"
	requiredOpenAPIRoute       = "/api"
	requiredOpenAPIArtifact    = "docs/generated/openapi.yaml"
	requiredRuntimeOpenAPIPath = "/openapi.yaml"
	requiredRuntimeDocsPath    = "/docs/api"
)

type scalarConfig struct {
	Scalar     string           `json:"scalar"`
	Info       scalarInfo       `json:"info"`
	Navigation scalarNavigation `json:"navigation"`
}

type scalarInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type scalarNavigation struct {
	Routes map[string]scalarRoute `json:"routes"`
}

type scalarRoute struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Filepath string                 `json:"filepath,omitempty"`
	Children map[string]scalarRoute `json:"children,omitempty"`
}

func main() {
	command := "check"
	if len(os.Args) > 1 {
		command = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}

	var err error
	switch command {
	case "check":
		err = runCheck()
	default:
		err = fmt.Errorf("unknown command %q (supported: check)", command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "scalar %s failed: %v\n", command, err)
		os.Exit(1)
	}

	fmt.Printf("scalar %s succeeded\n", command)
}

func runCheck() error {
	config, err := readConfig(configPath)
	if err != nil {
		return err
	}

	if strings.TrimSpace(config.Scalar) != requiredScalarVersion {
		return fmt.Errorf("scalar.config.json must declare scalar=%q", requiredScalarVersion)
	}
	if strings.TrimSpace(config.Info.Title) == "" {
		return errors.New("scalar.config.json info.title is required")
	}
	if strings.TrimSpace(config.Info.Description) == "" {
		return errors.New("scalar.config.json info.description is required")
	}
	if len(config.Navigation.Routes) == 0 {
		return errors.New("scalar.config.json navigation.routes must not be empty")
	}

	collectedPaths := make([]string, 0, 32)
	openAPIRoute, found := config.Navigation.Routes[requiredOpenAPIRoute]
	if !found {
		return fmt.Errorf("scalar.config.json must include %q route", requiredOpenAPIRoute)
	}
	if openAPIRoute.Type != "openapi" {
		return fmt.Errorf("route %q must use type=openapi", requiredOpenAPIRoute)
	}
	if strings.TrimSpace(openAPIRoute.Filepath) != requiredOpenAPIArtifact {
		return fmt.Errorf("route %q must reference %q", requiredOpenAPIRoute, requiredOpenAPIArtifact)
	}

	for routePath, route := range config.Navigation.Routes {
		paths, err := collectFilepaths(routePath, route)
		if err != nil {
			return err
		}
		collectedPaths = append(collectedPaths, paths...)
	}

	for _, path := range collectedPaths {
		if err := ensureReadableFile(path); err != nil {
			return err
		}
	}

	if err := ensureReadableFile(requiredOpenAPIArtifact); err != nil {
		return err
	}

	if err := ensureOpenAPIIncludesDocsRoutes(requiredOpenAPIArtifact); err != nil {
		return err
	}

	return nil
}

func readConfig(path string) (scalarConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return scalarConfig{}, fmt.Errorf("read %s: %w", path, err)
	}

	var cfg scalarConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return scalarConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}

	return cfg, nil
}

func collectFilepaths(routePath string, route scalarRoute) ([]string, error) {
	routeType := strings.TrimSpace(route.Type)
	switch routeType {
	case "page", "openapi":
		if strings.TrimSpace(route.Filepath) == "" {
			return nil, fmt.Errorf("route %q missing filepath", routePath)
		}
		return []string{route.Filepath}, nil
	case "group":
		if len(route.Children) == 0 {
			return nil, fmt.Errorf("group route %q must define children", routePath)
		}

		paths := make([]string, 0, len(route.Children))
		for childPath, child := range route.Children {
			childFilepaths, err := collectFilepaths(childPath, child)
			if err != nil {
				return nil, err
			}
			paths = append(paths, childFilepaths...)
		}
		return paths, nil
	default:
		return nil, fmt.Errorf("route %q has unsupported type %q", routePath, route.Type)
	}
}

func ensureReadableFile(path string) error {
	cleanPath := filepath.Clean(path)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("file path is not readable: %s (%w)", cleanPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("file path must be a regular file: %s", cleanPath)
	}
	return nil
}

func ensureOpenAPIIncludesDocsRoutes(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read openapi artifact %s: %w", path, err)
	}

	text := string(content)
	if !strings.Contains(text, requiredRuntimeOpenAPIPath+":") {
		return fmt.Errorf("openapi artifact %s must include runtime route %s", path, requiredRuntimeOpenAPIPath)
	}
	if !strings.Contains(text, requiredRuntimeDocsPath+":") {
		return fmt.Errorf("openapi artifact %s must include runtime route %s", path, requiredRuntimeDocsPath)
	}

	return nil
}
