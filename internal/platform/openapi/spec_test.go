package openapi

import (
	"path/filepath"
	"testing"
)

func TestLoadAndValidate_ReadSpecBytes_SpecRouteSet_SortedRouteKeys(t *testing.T) {
	specPath := filepath.Join("..", "..", "..", "api", "openapi", "openapi.yaml")

	doc, err := LoadAndValidate(specPath)
	if err != nil {
		t.Fatalf("load and validate spec: %v", err)
	}

	bytes, err := ReadSpecBytes(specPath)
	if err != nil {
		t.Fatalf("read spec bytes: %v", err)
	}
	if len(bytes) == 0 {
		t.Fatal("expected spec bytes not empty")
	}

	set := SpecRouteSet(doc)
	if len(set) == 0 {
		t.Fatal("expected route set not empty")
	}

	keys := SortedRouteKeys(set)
	if len(keys) == 0 {
		t.Fatal("expected sorted route keys not empty")
	}

	first := keys[0].String()
	if first == "" {
		t.Fatal("expected non-empty route key string")
	}
}
