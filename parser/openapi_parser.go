package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/holdemlab/ui-json-schema/schema"
)

// ErrInvalidOpenAPI is returned when the input is not a valid OpenAPI document.
var ErrInvalidOpenAPI = errors.New("invalid OpenAPI document")

// ErrSchemaNotFound is returned when the requested schema name is not found in the OpenAPI document.
var ErrSchemaNotFound = errors.New("schema not found in OpenAPI document")

// openAPIDoc is a minimal representation of an OpenAPI 3.x document
// containing only the parts relevant for schema extraction.
type openAPIDoc struct {
	Components struct {
		Schemas map[string]json.RawMessage `json:"schemas"`
	} `json:"components"`
}

// openAPISchema is a simplified OpenAPI/JSON-Schema object.
type openAPISchema struct {
	Type                 string                    `json:"type"`
	Properties           map[string]*openAPISchema `json:"properties,omitempty"`
	Items                *openAPISchema            `json:"items,omitempty"`
	AdditionalProperties *openAPISchema            `json:"additionalProperties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Format               string                    `json:"format,omitempty"`
	Default              any                       `json:"default,omitempty"`
	Enum                 []any                     `json:"enum,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Title                string                    `json:"title,omitempty"`
	Ref                  string                    `json:"$ref,omitempty"`
	AllSchemas           map[string]*openAPISchema `json:"-"`
}

// GenerateFromOpenAPI parses an OpenAPI 3.x JSON document and generates
// both a JSON Schema and a UI Schema for the named schema component.
// The schemaName must match a key under components.schemas.
func GenerateFromOpenAPI(data []byte, schemaName string) (*schema.JSONSchema, *schema.UISchemaElement, error) {
	var doc openAPIDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidOpenAPI, err) //nolint:errorlint // wrapping intentional
	}

	if len(doc.Components.Schemas) == 0 {
		return nil, nil, fmt.Errorf("%w: no components.schemas found", ErrInvalidOpenAPI)
	}

	raw, ok := doc.Components.Schemas[schemaName]
	if !ok {
		return nil, nil, fmt.Errorf("%w: %q", ErrSchemaNotFound, schemaName)
	}

	// Parse all schemas for $ref resolution.
	allSchemas := make(map[string]*openAPISchema)
	for name, rawSchema := range doc.Components.Schemas {
		var s openAPISchema
		if err := json.Unmarshal(rawSchema, &s); err != nil {
			continue
		}

		allSchemas[name] = &s
	}

	var oaSchema openAPISchema
	if err := json.Unmarshal(raw, &oaSchema); err != nil {
		return nil, nil, fmt.Errorf("%w: cannot parse schema %q: %v", ErrInvalidOpenAPI, schemaName, err) //nolint:errorlint // wrapping intentional
	}

	oaSchema.AllSchemas = allSchemas

	jsonSchema := convertOpenAPIToJSONSchema(&oaSchema, allSchemas)
	jsonSchema.Schema = schema.DefaultOptions().DraftURL()

	uiSchema := buildOpenAPIUISchema(&oaSchema, "#/properties", allSchemas)

	return jsonSchema, uiSchema, nil
}

// convertOpenAPIToJSONSchema converts an OpenAPI schema to a JSONSchema.
func convertOpenAPIToJSONSchema(oa *openAPISchema, allSchemas map[string]*openAPISchema) *schema.JSONSchema {
	// Resolve $ref.
	if oa.Ref != "" {
		resolved := resolveRef(oa.Ref, allSchemas)
		if resolved != nil {
			return convertOpenAPIToJSONSchema(resolved, allSchemas)
		}

		return &schema.JSONSchema{Type: "object"}
	}

	s := &schema.JSONSchema{
		Type:        oa.Type,
		Format:      oa.Format,
		Default:     oa.Default,
		Description: oa.Description,
		Title:       oa.Title,
	}

	if len(oa.Enum) > 0 {
		s.Enum = oa.Enum
	}

	if len(oa.Required) > 0 {
		s.Required = oa.Required
	}

	if len(oa.Properties) > 0 {
		s.Properties = make(map[string]*schema.JSONSchema)
		for name, prop := range oa.Properties {
			s.Properties[name] = convertOpenAPIToJSONSchema(prop, allSchemas)
		}
	}

	if oa.Items != nil {
		s.Items = convertOpenAPIToJSONSchema(oa.Items, allSchemas)
	}

	if oa.AdditionalProperties != nil {
		s.AdditionalProperties = convertOpenAPIToJSONSchema(oa.AdditionalProperties, allSchemas)
	}

	return s
}

// buildOpenAPIUISchema builds a UI Schema from an OpenAPI schema.
func buildOpenAPIUISchema(oa *openAPISchema, basePath string, allSchemas map[string]*openAPISchema) *schema.UISchemaElement {
	// Resolve $ref.
	if oa.Ref != "" {
		resolved := resolveRef(oa.Ref, allSchemas)
		if resolved != nil {
			return buildOpenAPIUISchema(resolved, basePath, allSchemas)
		}
	}

	root := schema.NewVerticalLayout()

	for name, prop := range oa.Properties {
		scope := basePath + "/" + name

		// Resolve $ref for property.
		actual := prop
		if prop.Ref != "" {
			if resolved := resolveRef(prop.Ref, allSchemas); resolved != nil {
				actual = resolved
			}
		}

		// Nested object → Group.
		if actual.Type == "object" && len(actual.Properties) > 0 {
			group := schema.NewGroup(name)
			nested := buildOpenAPIUISchema(actual, scope+"/properties", allSchemas)
			group.Elements = nested.Elements
			root.Elements = append(root.Elements, group)

			continue
		}

		control := schema.NewControl(scope)
		root.Elements = append(root.Elements, control)
	}

	return root
}

// resolveRef resolves a simple $ref of the form "#/components/schemas/Name".
func resolveRef(ref string, allSchemas map[string]*openAPISchema) *openAPISchema {
	const prefix = "#/components/schemas/"
	if len(ref) <= len(prefix) {
		return nil
	}

	name := ref[len(prefix):]

	return allSchemas[name]
}
