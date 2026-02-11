package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/holdemlab/ui-json-schema/parser"
	"github.com/holdemlab/ui-json-schema/schema"
)

// --- i18n ---

type I18nUser struct {
	Name  string `json:"name" i18n:"user.name" form:"label=Name"`
	Email string `json:"email" i18n:"user.email"`
}

func TestGenerateUISchemaWithOptions_I18n(t *testing.T) {
	tr := schema.NewMapTranslator(map[string]map[string]string{
		"uk": {
			"user.name":  "Ім'я",
			"user.email": "Електронна пошта",
		},
	})

	opts := schema.Options{
		Translator: tr,
		Locale:     "uk",
		Draft:      "draft-07",
	}

	ui, err := parser.GenerateUISchemaWithOptions(I18nUser{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	nameEl := ui.Elements[0]
	if nameEl.Label != "Ім'я" {
		t.Errorf("expected translated label, got %q", nameEl.Label)
	}

	emailEl := ui.Elements[1]
	if emailEl.Label != "Електронна пошта" {
		t.Errorf("expected translated label, got %q", emailEl.Label)
	}
}

func TestGenerateUISchemaWithOptions_I18n_NoTranslator(t *testing.T) {
	opts := schema.Options{Draft: "draft-07"}

	ui, err := parser.GenerateUISchemaWithOptions(I18nUser{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nameEl := ui.Elements[0]
	if nameEl.Label != "Name" {
		t.Errorf("expected form label 'Name', got %q", nameEl.Label)
	}

	emailEl := ui.Elements[1]
	if emailEl.Label != "user.email" {
		t.Errorf("expected i18n key as fallback label, got %q", emailEl.Label)
	}
}

func TestGenerateUISchemaWithOptions_I18n_MissingLocale(t *testing.T) {
	tr := schema.NewMapTranslator(map[string]map[string]string{
		"uk": {"user.name": "Ім'я"},
	})

	opts := schema.Options{
		Translator: tr,
		Locale:     "de",
		Draft:      "draft-07",
	}

	ui, err := parser.GenerateUISchemaWithOptions(I18nUser{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nameEl := ui.Elements[0]
	if nameEl.Label != "user.name" {
		t.Errorf("expected fallback to key, got %q", nameEl.Label)
	}
}

// --- Custom renderers ---

type RendererStruct struct {
	Color  string `json:"color" renderer:"color-picker"`
	Rating int    `json:"rating"`
}

func TestGenerateUISchemaWithOptions_RendererTag(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(RendererStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	colorEl := ui.Elements[0]
	if colorEl.Options == nil {
		t.Fatal("expected options on color")
	}

	if colorEl.Options["renderer"] != "color-picker" {
		t.Errorf("expected renderer 'color-picker', got %v", colorEl.Options["renderer"])
	}

	ratingEl := ui.Elements[1]
	if ratingEl.Options != nil {
		t.Errorf("expected nil options on rating, got %v", ratingEl.Options)
	}
}

func TestGenerateUISchemaWithOptions_RendererFromOptions(t *testing.T) {
	opts := schema.Options{
		Draft: "draft-07",
		Renderers: map[string]string{
			"#/properties/rating": "star-rating",
		},
	}

	ui, err := parser.GenerateUISchemaWithOptions(RendererStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ratingEl := ui.Elements[1]
	if ratingEl.Options == nil {
		t.Fatal("expected options on rating")
	}

	if ratingEl.Options["renderer"] != "star-rating" {
		t.Errorf("expected renderer 'star-rating', got %v", ratingEl.Options["renderer"])
	}
}

func TestGenerateUISchemaWithOptions_RendererTagOverridesOptions(t *testing.T) {
	opts := schema.Options{
		Draft: "draft-07",
		Renderers: map[string]string{
			"#/properties/color": "should-be-overridden",
		},
	}

	ui, err := parser.GenerateUISchemaWithOptions(RendererStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	colorEl := ui.Elements[0]
	if colorEl.Options["renderer"] != "color-picker" {
		t.Errorf("expected tag renderer 'color-picker', got %v", colorEl.Options["renderer"])
	}
}

// --- Permissions / readonly by role ---

type PermStruct struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func TestGenerateUISchemaWithOptions_RoleReadOnly(t *testing.T) {
	opts := schema.Options{
		Draft: "draft-07",
		Role:  "viewer",
		RolePermissions: map[string]schema.FieldPermissions{
			"viewer": {
				"name":  schema.AccessReadOnly,
				"email": schema.AccessReadOnly,
			},
		},
	}

	ui, err := parser.GenerateUISchemaWithOptions(PermStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	elemMap := make(map[string]*schema.UISchemaElement)
	for _, el := range ui.Elements {
		elemMap[el.Scope] = el
	}

	nameEl := elemMap["#/properties/name"]
	if nameEl == nil {
		t.Fatal("missing name element")
	}

	if nameEl.Options == nil || nameEl.Options["readonly"] != true {
		t.Errorf("expected name readonly for viewer, got %v", nameEl.Options)
	}

	emailEl := elemMap["#/properties/email"]
	if emailEl == nil {
		t.Fatal("missing email element")
	}

	if emailEl.Options == nil || emailEl.Options["readonly"] != true {
		t.Errorf("expected email readonly for viewer, got %v", emailEl.Options)
	}

	idEl := elemMap["#/properties/id"]
	if idEl != nil && idEl.Options != nil && idEl.Options["readonly"] == true {
		t.Error("id should not be readonly")
	}
}

func TestGenerateUISchemaWithOptions_RoleHidden(t *testing.T) {
	opts := schema.Options{
		Draft: "draft-07",
		Role:  "viewer",
		RolePermissions: map[string]schema.FieldPermissions{
			"viewer": {
				"role": schema.AccessHidden,
			},
		},
	}

	ui, err := parser.GenerateUISchemaWithOptions(PermStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, el := range ui.Elements {
		if el.Scope == "#/properties/role" {
			t.Error("role field should be hidden for viewer")
		}
	}

	if len(ui.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(ui.Elements))
	}
}

func TestGenerateUISchemaWithOptions_NoRole(t *testing.T) {
	opts := schema.Options{
		Draft: "draft-07",
		RolePermissions: map[string]schema.FieldPermissions{
			"viewer": {
				"role": schema.AccessHidden,
			},
		},
	}

	ui, err := parser.GenerateUISchemaWithOptions(PermStruct{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 4 {
		t.Errorf("expected 4 elements (no role active), got %d", len(ui.Elements))
	}
}

// --- Draft 2019-09 ---

func TestGenerateJSONSchemaWithOptions_Draft2019(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	opts := schema.Options{Draft: "2019-09"}

	s, err := parser.GenerateJSONSchemaWithOptions(Simple{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://json-schema.org/draft/2019-09/schema"
	if s.Schema != expected {
		t.Errorf("expected $schema %q, got %q", expected, s.Schema)
	}
}

func TestGenerateJSONSchemaWithOptions_Draft07(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
	}

	opts := schema.Options{Draft: "draft-07"}

	s, err := parser.GenerateJSONSchemaWithOptions(Simple{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Schema != schemaDraft7 {
		t.Errorf("expected $schema %q, got %q", schemaDraft7, s.Schema)
	}
}

func TestGenerateFromJSONWithOptions_Draft2019(t *testing.T) {
	input := []byte(`{"name": "John"}`)
	opts := schema.Options{Draft: "2019-09"}

	s, _, err := parser.GenerateFromJSONWithOptions(input, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://json-schema.org/draft/2019-09/schema"
	if s.Schema != expected {
		t.Errorf("expected $schema %q, got %q", expected, s.Schema)
	}
}

// --- Custom layouts: Categorization ---

type CategorizedForm struct {
	Name  string `json:"name" form:"category=Personal"`
	Email string `json:"email" form:"category=Personal"`
	Role  string `json:"role" form:"category=Work"`
	Bio   string `json:"bio" form:"category=Work;multiline"`
}

func TestGenerateUISchemaWithOptions_Categorization(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategorizedForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization root, got %q", ui.Type)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(ui.Elements))
	}

	personal := ui.Elements[0]
	if personal.Type != "Category" {
		t.Errorf("expected Category type, got %q", personal.Type)
	}

	if personal.Label != "Personal" {
		t.Errorf("expected label 'Personal', got %q", personal.Label)
	}

	if len(personal.Elements) != 2 {
		t.Errorf("expected 2 elements in Personal, got %d", len(personal.Elements))
	}

	work := ui.Elements[1]
	if work.Label != "Work" {
		t.Errorf("expected label 'Work', got %q", work.Label)
	}

	if len(work.Elements) != 2 {
		t.Errorf("expected 2 elements in Work, got %d", len(work.Elements))
	}

	bioEl := work.Elements[1]
	if bioEl.Options == nil || bioEl.Options["multi"] != true {
		t.Errorf("expected multi=true on bio, got %v", bioEl.Options)
	}
}

type MixedCategoryForm struct {
	Name  string `json:"name" form:"category=Personal"`
	Email string `json:"email"`
}

func TestGenerateUISchemaWithOptions_MixedCategories(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(MixedCategoryForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization root, got %q", ui.Type)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(ui.Elements))
	}

	if ui.Elements[0].Label != "Personal" {
		t.Errorf("expected first category 'Personal', got %q", ui.Elements[0].Label)
	}

	if ui.Elements[1].Label != "Other" {
		t.Errorf("expected second category 'Other', got %q", ui.Elements[1].Label)
	}
}

type NoCategoryForm struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestGenerateUISchemaWithOptions_NoCategories(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(NoCategoryForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}
}

// --- OpenAPI ---

func TestGenerateFromOpenAPI_Simple(t *testing.T) {
	doc := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"age": {"type": "integer"},
						"email": {"type": "string", "format": "email"}
					},
					"required": ["name", "email"]
				}
			}
		}
	}`

	s, ui, err := parser.GenerateFromOpenAPI([]byte(doc), "User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Schema != schemaDraft7 {
		t.Errorf("expected $schema Draft 7, got %q", s.Schema)
	}

	if s.Type != typeObject {
		t.Errorf("expected type object, got %q", s.Type)
	}

	if len(s.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(s.Properties))
	}

	nameProp := s.Properties["name"]
	if nameProp == nil || nameProp.Type != typeString {
		t.Errorf("expected name string, got %v", nameProp)
	}

	emailProp := s.Properties["email"]
	if emailProp == nil || emailProp.Format != formatEmail {
		t.Errorf("expected email format, got %v", emailProp)
	}

	if len(s.Required) != 2 {
		t.Errorf("expected 2 required, got %d", len(s.Required))
	}

	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 3 {
		t.Errorf("expected 3 UI elements, got %d", len(ui.Elements))
	}
}

func TestGenerateFromOpenAPI_NestedObject(t *testing.T) {
	doc := `{
		"components": {
			"schemas": {
				"Order": {
					"type": "object",
					"properties": {
						"id": {"type": "integer"},
						"address": {
							"type": "object",
							"properties": {
								"city": {"type": "string"},
								"zip": {"type": "string"}
							}
						}
					}
				}
			}
		}
	}`

	s, ui, err := parser.GenerateFromOpenAPI([]byte(doc), "Order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	addrProp := s.Properties["address"]
	if addrProp == nil || addrProp.Type != typeObject {
		t.Fatalf("expected address object, got %v", addrProp)
	}

	if len(addrProp.Properties) != 2 {
		t.Errorf("expected 2 addr props, got %d", len(addrProp.Properties))
	}

	var addressGroup bool
	for _, el := range ui.Elements {
		if el.Type == typeGroup {
			addressGroup = true
		}
	}

	if !addressGroup {
		t.Error("expected address Group in UI Schema")
	}
}

func TestGenerateFromOpenAPI_Ref(t *testing.T) {
	doc := `{
		"components": {
			"schemas": {
				"Address": {
					"type": "object",
					"properties": {
						"city": {"type": "string"},
						"zip": {"type": "string"}
					}
				},
				"User": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"address": {"$ref": "#/components/schemas/Address"}
					}
				}
			}
		}
	}`

	s, ui, err := parser.GenerateFromOpenAPI([]byte(doc), "User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	addrProp := s.Properties["address"]
	if addrProp == nil || addrProp.Type != typeObject {
		t.Fatalf("expected address resolved to object, got %v", addrProp)
	}

	if len(addrProp.Properties) != 2 {
		t.Errorf("expected 2 addr props from ref, got %d", len(addrProp.Properties))
	}

	var hasGroup bool
	for _, el := range ui.Elements {
		if el.Type == typeGroup {
			hasGroup = true
		}
	}

	if !hasGroup {
		t.Error("expected Group for resolved $ref address")
	}
}

func TestGenerateFromOpenAPI_SchemaNotFound(t *testing.T) {
	doc := `{"components": {"schemas": {"User": {"type": "object"}}}}`

	_, _, err := parser.GenerateFromOpenAPI([]byte(doc), "Missing")
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

func TestGenerateFromOpenAPI_InvalidJSON(t *testing.T) {
	_, _, err := parser.GenerateFromOpenAPI([]byte("not json"), "User")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGenerateFromOpenAPI_NoSchemas(t *testing.T) {
	doc := `{"components": {}}`

	_, _, err := parser.GenerateFromOpenAPI([]byte(doc), "User")
	if err == nil {
		t.Fatal("expected error for empty schemas")
	}
}

func TestGenerateFromOpenAPI_WithEnum(t *testing.T) {
	doc := `{
		"components": {
			"schemas": {
				"Config": {
					"type": "object",
					"properties": {
						"env": {
							"type": "string",
							"enum": ["dev", "staging", "prod"]
						}
					}
				}
			}
		}
	}`

	s, _, err := parser.GenerateFromOpenAPI([]byte(doc), "Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envProp := s.Properties["env"]
	if envProp == nil {
		t.Fatal("missing env property")
	}

	if len(envProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(envProp.Enum))
	}
}

func TestGenerateFromOpenAPI_ArrayItems(t *testing.T) {
	doc := `{
		"components": {
			"schemas": {
				"TagList": {
					"type": "object",
					"properties": {
						"tags": {
							"type": "array",
							"items": {"type": "string"}
						}
					}
				}
			}
		}
	}`

	s, _, err := parser.GenerateFromOpenAPI([]byte(doc), "TagList")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tagsProp := s.Properties["tags"]
	if tagsProp == nil || tagsProp.Type != "array" {
		t.Fatal("expected tags as array")
	}

	if tagsProp.Items == nil || tagsProp.Items.Type != typeString {
		t.Errorf("expected items string, got %v", tagsProp.Items)
	}
}

// --- JSON serialization ---

func TestCategorization_JSON(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategorizedForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(ui)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if raw["type"] != typeCategorization {
		t.Errorf("expected Categorization in JSON, got %v", raw["type"])
	}

	elements, ok := raw["elements"].([]any)
	if !ok {
		t.Fatal("missing elements")
	}

	if len(elements) != 2 {
		t.Errorf("expected 2 categories, got %d", len(elements))
	}

	cat1 := elements[0].(map[string]any)
	if cat1["type"] != "Category" {
		t.Errorf("expected Category type, got %v", cat1["type"])
	}
}
