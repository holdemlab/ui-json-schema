package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/holdemlab/ui-json-schema/schema"
)

// ErrInvalidJSON is returned when the input data is not valid JSON.
var ErrInvalidJSON = errors.New("invalid JSON")

// ErrNotJSONObject is returned when the top-level JSON value is not an object.
var ErrNotJSONObject = errors.New("top-level JSON value must be an object")

// GenerateFromJSON generates both a JSON Schema and a UI Schema from raw JSON bytes.
// The input must be a JSON object (not an array or primitive).
// All fields are treated as optional (no required).
func GenerateFromJSON(data []byte) (*schema.JSONSchema, *schema.UISchemaElement, error) {
	return GenerateFromJSONWithOptions(data, schema.DefaultOptions())
}

// GenerateFromJSONWithOptions generates both schemas from raw JSON bytes using the supplied options.
func GenerateFromJSONWithOptions(data []byte, opts schema.Options) (*schema.JSONSchema, *schema.UISchemaElement, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err) //nolint:errorlint // wrapping intentional
	}

	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, nil, ErrNotJSONObject
	}

	root := &schema.JSONSchema{
		Schema: opts.DraftURL(),
		Type:   "object",
	}
	root.Properties = make(map[string]*schema.JSONSchema)

	uiRoot := schema.NewVerticalLayout()

	buildFromObject(obj, root, "#/properties", uiRoot)

	return root, uiRoot, nil
}

// buildFromObject iterates over a JSON object and populates JSON Schema properties
// and UI Schema elements.
func buildFromObject(obj map[string]any, s *schema.JSONSchema, basePath string, parent *schema.UISchemaElement) {
	for key, val := range obj {
		prop := inferSchema(val)
		s.Properties[key] = prop

		scope := basePath + "/" + key

		// Nested objects get a Group layout.
		if nested, ok := val.(map[string]any); ok {
			group := schema.NewGroup(key)
			buildFromObject(nested, prop, scope+"/properties", group)
			parent.Elements = append(parent.Elements, group)

			continue
		}

		control := schema.NewControl(scope)
		parent.Elements = append(parent.Elements, control)
	}
}

// inferSchema infers a JSONSchema from an arbitrary JSON value.
func inferSchema(val any) *schema.JSONSchema {
	switch v := val.(type) {
	case nil:
		return &schema.JSONSchema{Type: "null"}

	case bool:
		return &schema.JSONSchema{Type: "boolean"}

	case float64:
		return inferNumberSchema(v)

	case string:
		return &schema.JSONSchema{Type: "string"}

	case []any:
		return inferArraySchema(v)

	case map[string]any:
		obj := &schema.JSONSchema{
			Type:       "object",
			Properties: make(map[string]*schema.JSONSchema),
		}
		// Properties are populated by buildFromObject at the call site.
		return obj

	default:
		return &schema.JSONSchema{Type: "string"}
	}
}

// inferNumberSchema determines whether a JSON number is an integer or a float.
func inferNumberSchema(v float64) *schema.JSONSchema {
	if v == math.Trunc(v) && !math.IsInf(v, 0) && !math.IsNaN(v) {
		return &schema.JSONSchema{Type: "integer"}
	}

	return &schema.JSONSchema{Type: "number"}
}

// inferArraySchema infers the items schema from a JSON array.
// If the array is empty, items is set to an empty schema.
// If the array has elements, the first element's type is used.
func inferArraySchema(arr []any) *schema.JSONSchema {
	s := &schema.JSONSchema{Type: "array"}

	if len(arr) == 0 {
		s.Items = &schema.JSONSchema{}
		return s
	}

	s.Items = inferSchema(arr[0])

	// If the first item is a nested object, populate its properties.
	if nested, ok := arr[0].(map[string]any); ok {
		for key, val := range nested {
			s.Items.Properties[key] = inferSchema(val)
		}
	}

	return s
}
