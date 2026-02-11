package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/holdemlab/ui-json-schema/schema"
)

func TestNewJSONSchema(t *testing.T) {
	s := schema.NewJSONSchema()

	if s.Schema != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("expected Draft-07 $schema, got %q", s.Schema)
	}

	if s.Type != "object" {
		t.Errorf("expected type 'object', got %q", s.Type)
	}
}

func TestJSONSchema_MarshalJSON(t *testing.T) {
	s := &schema.JSONSchema{
		Type: "object",
		Properties: map[string]*schema.JSONSchema{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result["type"] != "object" {
		t.Errorf("expected type 'object', got %v", result["type"])
	}

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	if len(props) != 2 {
		t.Errorf("expected 2 properties, got %d", len(props))
	}
}

func TestJSONSchema_OmitEmpty(t *testing.T) {
	s := &schema.JSONSchema{
		Type: "string",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	for _, field := range []string{"$schema", "properties", "items", "required", "format", "enum", "description", "title"} {
		if _, ok := result[field]; ok {
			t.Errorf("expected field %q to be omitted", field)
		}
	}
}
