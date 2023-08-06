package main

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"os"
	"strings"
)

type Schema struct {
	Title            string                 `json:"title"`
	Type             string                 `json:"type"`
	Format           string                 `json:"format,omitempty"`
	Minimum          float64                `json:"minimum,omitempty"`
	Maximum          float64                `json:"maximum,omitempty"`
	MinItems         int                    `json:"minItems,omitempty"`
	MaxItems         int                    `json:"maxItems,omitempty"`
	Properties       map[string]interface{} `json:"properties"`
	Definitions      map[string]Schema      `json:"definitions"`
	Ref              string                 `json:"$ref"`
	Items            *Schema                `json:"items"`
	pathToDirSchemas string
}


func getSchemaFromFile(filename string) (Schema, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return Schema{}, err
	}

	var schema Schema
	err = json.Unmarshal(b, &schema)
	if err != nil {
		return Schema{}, err
	}

	return schema, nil
}

func ResolveRefs(rootSchema Schema, schema Schema) (Schema, error) {
	if schema.Ref != "" {
		if strings.HasPrefix(schema.Ref, "#") {
			refPath := schema.Ref[2:]
			refParts := strings.Split(refPath, "/")
			definition := rootSchema.Definitions[refParts[1]]
			return ResolveRefs(rootSchema, definition)
		} else {
			// Обработка ссылки на внешний файл
			filename := schema.Ref
			externalSchema, err := getSchemaFromFile(rootSchema.pathToDirSchemas + "/" + filename)
			if err != nil {
				return Schema{}, err
			}
			return ResolveRefs(externalSchema, externalSchema)
		}
	}

	for propName, propValue := range schema.Properties {
		if propSchema, ok := propValue.(map[string]interface{}); ok {
			var propSchemaStruct Schema
			mapToStruct(propSchema, &propSchemaStruct)
			resolvedSchema, err := ResolveRefs(rootSchema, propSchemaStruct)
			if err != nil {
				return Schema{}, err
			}
			schema.Properties[propName] = resolvedSchema
		}
	}

	return schema, nil
}

func mapToStruct(m map[string]interface{}, s *Schema) {
	j, _ := json.Marshal(m)
	json.Unmarshal(j, s)
}
