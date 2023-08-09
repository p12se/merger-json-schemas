package merger

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	Title = "title"
)

type JSONSchema map[string]interface{}

type SchemaLoader interface {
	Load(ref string) (JSONSchema, error)
}

func MergeSchemas(target, source JSONSchema) {
	for k, v := range source {
		if !strings.EqualFold(k, Title) {
			target[k] = v
		}
	}
}

type FileSchemaLoader struct {
	BaseDir string
}

func (f *FileSchemaLoader) Load(ref string) (JSONSchema, error) {
	fullPath := path.Join(f.BaseDir, ref)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var schema JSONSchema
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

type HttpSchemaLoader struct {
	BaseURL string
}

func (h *HttpSchemaLoader) Load(ref string) (*JSONSchema, error) {
	res, err := http.Get(h.BaseURL + "/" + ref)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var schema JSONSchema
	err = json.NewDecoder(res.Body).Decode(&schema)
	return &schema, err
}

func ResolveRefs(schema JSONSchema, loader SchemaLoader, currentPath string) error {
	if ref, ok := schema["$ref"].(string); ok {
		relativePath := path.Join(path.Dir(currentPath), ref)

		refSchema, err := loader.Load(relativePath)
		if err != nil {
			return err
		}
		MergeSchemas(schema, refSchema)
		delete(schema, "$ref")

		currentPath = relativePath
	}

	for _, v := range schema {
		if subSchema, ok := v.(map[string]interface{}); ok {
			if err := ResolveRefs(subSchema, loader, currentPath); err != nil {
				return err
			}
		}
	}
	return nil
}
