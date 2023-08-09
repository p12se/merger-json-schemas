package merger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileSchemaLoader_Load(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "./testdata"}

	_, err := loader.Load("example.json")
	if err != nil {
		t.Fatalf("Failed to load schema: %s", err)
	}
}

func TestFileSchemaLoader_FileNotFound(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "./testdata"}

	_, err := loader.Load("non_existent.json")
	if err == nil {
		t.Fatalf("Expected an error for non-existent file")
	}
}

func TestHttpSchemaLoader_Load(t *testing.T) {
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

	if value, ok := (*schema)["type"].(string); !ok || value != "string" {
		t.Errorf("Expected type to be 'string', got %s", value)
	}
}

func TestHttpSchemaLoader_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	loader := &HttpSchemaLoader{BaseURL: server.URL}
	_, err := loader.Load("/non_existent.json")
	if err == nil {
		t.Fatalf("Expected an error for non-existent endpoint")
	}
}

func TestResolveRefs_SimpleRef(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "./testdata"}
	schema := JSONSchema{
		"$ref": "example.json",
	}

	err := ResolveRefs(schema, loader, "main.json")
	if err != nil {
		t.Fatalf("Failed to resolve refs: %s", err)
	}

	if _, exists := schema["$ref"]; exists {
		t.Errorf("$ref should be removed after being resolved")
	}
}

func TestResolveRefs_NestedRefs(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "./testdata"}
	schema := JSONSchema{
		"properties": map[string]interface{}{
			"user": map[string]interface{}{
				"$ref": "user.json",
			},
		},
	}

	err := ResolveRefs(schema, loader, "main.json")
	if err != nil {
		t.Fatalf("Failed to resolve refs: %s", err)
	}

	userProp, ok := schema["properties"].(map[string]interface{})["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'user' to be an object after resolving refs")
	}

	if _, exists := userProp["$ref"]; exists {
		t.Errorf("Nested $ref should be removed after being resolved")
	}
}

func TestResolveRefs_RelativePath(t *testing.T) {
	loader := &FileSchemaLoader{BaseDir: "./testdata"}
	schema := JSONSchema{
		"$ref": "./nested/folder/example.json",
	}

	err := ResolveRefs(schema, loader, "main.json")
	if err != nil {
		t.Fatalf("Failed to resolve refs: %s", err)
	}

	if _, exists := schema["$ref"]; exists {
		t.Errorf("$ref should be removed after being resolved")
	}
}
