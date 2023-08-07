package merger

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
)

type SchemaLoader interface {
	Load(ref string) (*JSONSchema, error)
}

type JSONSchema struct {
	Title            string                 `json:"title"`
	Type             string                 `json:"type"`
	Format           string                 `json:"format,omitempty"`
	Minimum          float64                `json:"minimum,omitempty"`
	Maximum          float64                `json:"maximum,omitempty"`
	MinItems         int                    `json:"minItems,omitempty"`
	MaxItems         int                    `json:"maxItems,omitempty"`
	Properties       map[string]interface{} `json:"properties"`
	Definitions      map[string]JSONSchema  `json:"definitions"`
	Ref              string                 `json:"$ref"`
	Items            *JSONSchema            `json:"items"`
	pathToDirSchemas string
}

type FileSchemaLoader struct {
	BaseDir string
}

func (f *FileSchemaLoader) Load(ref string) (*JSONSchema, error) {
	bytes, err := os.ReadFile(path.Join(f.BaseDir, ref))
	if err != nil {
		return nil, err
	}
	var schema JSONSchema
	err = json.Unmarshal(bytes, &schema)
	return &schema, err
}

type HttpSchemaLoader struct {
	BaseURL string
}

func (h *HttpSchemaLoader) Load(ref string) (*JSONSchema, error) {
	res, err := http.Get(path.Join(h.BaseURL, ref))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var schema JSONSchema
	err = json.NewDecoder(res.Body).Decode(&schema)
	return &schema, err
}

func getSchemaFromFile(filename string) (JSONSchema, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return JSONSchema{}, err
	}

	var schema JSONSchema
	err = json.Unmarshal(b, &schema)
	if err != nil {
		return JSONSchema{}, err
	}

	return schema, nil
}

func ResolveRefs(schema *JSONSchema, loader SchemaLoader) error {
	if schema.Ref != "" {
		refSchema, err := loader.Load(schema.Ref)
		if err != nil {
			return err
		}
		if err := ResolveRefs(refSchema, loader); err != nil {
			return err
		}
		*schema = *refSchema
	}

	for _, prop := range schema.Properties {
		if nestedSchema, ok := prop.(JSONSchema); ok {
			if err := ResolveRefs(&nestedSchema, loader); err != nil {
				return err
			}
		}
	}

	return nil
}
