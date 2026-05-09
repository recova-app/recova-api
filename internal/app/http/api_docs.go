package http

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	openAPIRoutePath        = "/openapi.yaml"
	scalarDocsRoutePath     = "/docs/api"
	openAPISourceFilePath   = "docs/generated/openapi.yaml"
	scalarScriptSource      = "https://cdn.jsdelivr.net/npm/@scalar/api-reference"
	scalarDocsPageTitle     = "Recova Backend API Reference"
	scalarDocsCSPTemplate   = "default-src 'none'; base-uri 'none'; frame-ancestors 'none'; connect-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; font-src 'self' https://fonts.scalar.com data:; script-src 'self' %s 'nonce-%s';"
	scalarDocsResponseCache = "no-store"
)

func (s *Server) registerAPIDocsRoutes() {
	s.app.Get(openAPIRoutePath, func(c fiber.Ctx) error {
		return serveOpenAPIArtifact(c)
	})
	s.app.Get(scalarDocsRoutePath, func(c fiber.Ctx) error {
		return serveScalarDocs(c)
	})
	s.app.Get(scalarDocsRoutePath+"/", func(c fiber.Ctx) error {
		return serveScalarDocs(c)
	})
}

func serveOpenAPIArtifact(c fiber.Ctx) error {
	specPath, err := resolveProjectPath(openAPISourceFilePath)
	if err != nil {
		return errs.New(
			errs.CodeServiceUnavailable,
			"Dokumen OpenAPI belum tersedia",
			fiber.Map{"path": openAPISourceFilePath},
			nil,
		)
	}

	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		return errs.New(
			errs.CodeServiceUnavailable,
			"Dokumen OpenAPI belum tersedia",
			fiber.Map{"path": openAPISourceFilePath},
			nil,
		)
	}

	c.Set("Content-Type", "application/yaml; charset=utf-8")
	c.Set("Cache-Control", scalarDocsResponseCache)
	return c.Send(specBytes)
}

func serveScalarDocs(c fiber.Ctx) error {
	nonce, err := makeNonce()
	if err != nil {
		return errs.New(errs.CodeInternalError, "Dokumentasi API gagal dimuat", nil, nil)
	}

	csp := fmt.Sprintf(scalarDocsCSPTemplate, scalarScriptSource, nonce)
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", scalarDocsResponseCache)
	c.Set("Content-Security-Policy", csp)

	return c.SendString(renderScalarHTML(nonce))
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

func makeNonce() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(raw), nil
}

func renderScalarHTML(nonce string) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n")
	b.WriteString("<html lang=\"en\">\n")
	b.WriteString("  <head>\n")
	b.WriteString("    <meta charset=\"utf-8\" />\n")
	b.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	b.WriteString("    <title>" + scalarDocsPageTitle + "</title>\n")
	b.WriteString("  </head>\n")
	b.WriteString("  <body>\n")
	b.WriteString("    <div id=\"app\"></div>\n")
	b.WriteString("    <script src=\"" + scalarScriptSource + "\"></script>\n")
	b.WriteString("    <script nonce=\"" + nonce + "\">\n")
	b.WriteString("      Scalar.createApiReference('#app', { url: '" + openAPIRoutePath + "', theme: 'default', layout: 'modern' })\n")
	b.WriteString("    </script>\n")
	b.WriteString("  </body>\n")
	b.WriteString("</html>\n")
	return b.String()
}
