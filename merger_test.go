package merger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileSchemaLoader(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "testdata"}

	schema, err := loader.Load("test1.json")
	if err != nil {
		t.Fatalf("Failed to load schema: %s", err)
	}

	if schema.Ref != "test2.json" {
		t.Errorf("Expected $ref to be 'test2.json', got %s", schema.Ref)
	}
}

func TestResolveRefs(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "testdata"}
	schema := &JSONSchema{
		Ref: "test1.json",
	}

	err := ResolveRefs(schema, loader)
	if err != nil {
		t.Fatalf("Failed to resolve refs: %s", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type to be 'string', got %s", schema.Type)
	}
}

func TestHttpSchemaLoader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test.json" {
			w.Write([]byte(`{"type": "string"}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	loader := &HttpSchemaLoader{BaseURL: server.URL}

	schema, err := loader.Load("test.json")
	if err != nil {
		t.Fatalf("Failed to load schema: %s", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type to be 'string', got %s", schema.Type)
	}
}
