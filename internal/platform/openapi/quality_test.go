package openapi

import (
	"context"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestValidateProductionReadiness_Pass(t *testing.T) {
	doc := mustLoadDoc(t, `
openapi: 3.1.0
info:
  title: test
  version: 1.0.0
paths:
  /hello:
    get:
      operationId: getHello
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
              examples:
                success:
                  value:
                    success: true
                    message: Data berhasil diambil
`)

	if err := ValidateProductionReadiness(doc); err != nil {
		t.Fatalf("expected pass, got err: %v", err)
	}
}

func TestValidateProductionReadiness_FailMissingExamples(t *testing.T) {
	doc := mustLoadDoc(t, `
openapi: 3.1.0
info:
  title: test
  version: 1.0.0
paths:
  /hello:
    get:
      operationId: getHello
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
`)

	err := ValidateProductionReadiness(doc)
	if err == nil {
		t.Fatal("expected error for missing examples")
	}
	if !strings.Contains(err.Error(), "missing examples") {
		t.Fatalf("expected missing examples error, got: %v", err)
	}
}

func TestValidateProductionReadiness_FailDuplicateOperationID(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: test
  version: 1.0.0
paths:
  /a:
    get:
      operationId: op
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema: { type: object }
              examples:
                ok:
                  value: { message: valid }
  /b:
    get:
      operationId: op
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema: { type: object }
              examples:
                ok:
                  value: { message: valid }
`))
	if err != nil {
		t.Fatalf("load doc: %v", err)
	}

	err = ValidateProductionReadiness(doc)
	if err == nil {
		t.Fatal("expected error for duplicate operationId")
	}
	if !strings.Contains(err.Error(), "duplicate operationId") {
		t.Fatalf("expected duplicate operationId error, got: %v", err)
	}
}

func TestValidateProductionReadiness_FailTemplatePlaceholder(t *testing.T) {
	doc := mustLoadDoc(t, `
openapi: 3.1.0
info:
  title: test
  version: 1.0.0
paths:
  /hello:
    get:
      operationId: getHello
      responses:
        '200':
          description: ok
          content:
            application/json:
              schema:
                type: object
              examples:
                bad:
                  value:
                    message: string
                    data:
                      additionalProperty: string
`)

	err := ValidateProductionReadiness(doc)
	if err == nil {
		t.Fatal("expected error for placeholder")
	}
	if !strings.Contains(err.Error(), "placeholder") {
		t.Fatalf("expected placeholder error, got: %v", err)
	}
}

func mustLoadDoc(t *testing.T, spec string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	if err != nil {
		t.Fatalf("load doc: %v", err)
	}
	if err := doc.Validate(context.Background(), openapi3.IsOpenAPI31OrLater()); err != nil {
		t.Fatalf("validate doc: %v", err)
	}
	return doc
}
