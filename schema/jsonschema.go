// Package schema provides types and functions for generating
// JSON Schema and UI Schema from Go structs and JSON objects.
package schema

// JSONSchema represents a JSON Schema document (Draft 7 compatible).
type JSONSchema struct {
	Schema               string                 `json:"$schema,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	AdditionalProperties *JSONSchema            `json:"additionalProperties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Default              any                    `json:"default,omitempty"`
	Enum                 []any                  `json:"enum,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Const                any                    `json:"const,omitempty"`
}

// NewJSONSchema creates a root JSON Schema object with the $schema field set.
func NewJSONSchema() *JSONSchema {
	return &JSONSchema{
		Schema: "http://json-schema.org/draft-07/schema#",
		Type:   "object",
	}
}
