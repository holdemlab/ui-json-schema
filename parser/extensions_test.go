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

// --- HorizontalLayout ---

type HorizontalPair struct {
	FirstName string `json:"firstName" form:"layout=horizontal"`
	LastName  string `json:"lastName" form:"layout=horizontal"`
	Email     string `json:"email"`
}

type HorizontalMixed struct {
	FirstName string `json:"firstName" form:"layout=horizontal"`
	LastName  string `json:"lastName" form:"layout=horizontal"`
	Bio       string `json:"bio"`
	City      string `json:"city" form:"layout=horizontal"`
	Country   string `json:"country" form:"layout=horizontal"`
}

type HorizontalSingle struct {
	Name  string `json:"name" form:"layout=horizontal"`
	Email string `json:"email"`
}

type HorizontalNested struct {
	Title   string `json:"title"`
	Address struct {
		Street string `json:"street" form:"layout=horizontal"`
		Number string `json:"number" form:"layout=horizontal"`
		City   string `json:"city"`
	} `json:"address"`
}

type HorizontalCategory struct {
	FirstName string `json:"firstName" form:"category=Personal;layout=horizontal"`
	LastName  string `json:"lastName" form:"category=Personal;layout=horizontal"`
	Email     string `json:"email" form:"category=Contact"`
	Phone     string `json:"phone" form:"category=Contact;layout=horizontal"`
	Fax       string `json:"fax" form:"category=Contact;layout=horizontal"`
}

type HorizontalAllFields struct {
	A string `json:"a" form:"layout=horizontal"`
	B string `json:"b" form:"layout=horizontal"`
	C string `json:"c" form:"layout=horizontal"`
}

func TestGenerateUISchema_HorizontalLayout_Pair(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalPair{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 elements: HorizontalLayout + Control(email)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	hl := ui.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Fatalf("expected 2 elements in HorizontalLayout, got %d", len(hl.Elements))
	}

	if hl.Elements[0].Scope != "#/properties/firstName" {
		t.Errorf("expected firstName scope, got %q", hl.Elements[0].Scope)
	}

	if hl.Elements[1].Scope != "#/properties/lastName" {
		t.Errorf("expected lastName scope, got %q", hl.Elements[1].Scope)
	}

	email := ui.Elements[1]
	if email.Type != "Control" {
		t.Errorf("expected Control, got %q", email.Type)
	}

	if email.Scope != "#/properties/email" {
		t.Errorf("expected email scope, got %q", email.Scope)
	}
}

func TestGenerateUISchema_HorizontalLayout_Mixed(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalMixed{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// HL(firstName,lastName) + Control(bio) + HL(city,country)
	if len(ui.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(ui.Elements))
	}

	hl1 := ui.Elements[0]
	if hl1.Type != "HorizontalLayout" {
		t.Errorf("expected first HorizontalLayout, got %q", hl1.Type)
	}

	if len(hl1.Elements) != 2 {
		t.Errorf("expected 2 elements in first HL, got %d", len(hl1.Elements))
	}

	bio := ui.Elements[1]
	if bio.Type != "Control" || bio.Scope != "#/properties/bio" {
		t.Errorf("expected bio Control, got %q %q", bio.Type, bio.Scope)
	}

	hl2 := ui.Elements[2]
	if hl2.Type != "HorizontalLayout" {
		t.Errorf("expected second HorizontalLayout, got %q", hl2.Type)
	}

	if len(hl2.Elements) != 2 {
		t.Errorf("expected 2 elements in second HL, got %d", len(hl2.Elements))
	}
}

func TestGenerateUISchema_HorizontalLayout_SingleField(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalSingle{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Single horizontal field → no HorizontalLayout created, just Controls.
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	for _, el := range ui.Elements {
		if el.Type != "Control" {
			t.Errorf("expected Control for single horizontal field, got %q", el.Type)
		}
	}
}

func TestGenerateUISchema_HorizontalLayout_Nested(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalNested{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Control(title) + Group(address)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 root elements, got %d", len(ui.Elements))
	}

	group := ui.Elements[1]
	if group.Type != typeGroup {
		t.Fatalf("expected Group, got %q", group.Type)
	}

	// Inside group: HL(street,number) + Control(city)
	if len(group.Elements) != 2 {
		t.Fatalf("expected 2 group elements, got %d", len(group.Elements))
	}

	hl := group.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout inside Group, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Errorf("expected 2 elements in nested HL, got %d", len(hl.Elements))
	}

	city := group.Elements[1]
	if city.Type != "Control" || city.Scope != "#/properties/address/properties/city" {
		t.Errorf("expected city Control, got %q %q", city.Type, city.Scope)
	}
}

func TestGenerateUISchema_HorizontalLayout_WithCategory(t *testing.T) {
	ui, err := parser.GenerateUISchemaWithOptions(HorizontalCategory{}, schema.DefaultOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(ui.Elements))
	}

	// Personal category: HL(firstName,lastName)
	personal := ui.Elements[0]
	if len(personal.Elements) != 1 {
		t.Fatalf("expected 1 element in Personal, got %d", len(personal.Elements))
	}

	hl := personal.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout in Personal, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Errorf("expected 2 controls in Personal HL, got %d", len(hl.Elements))
	}

	// Contact category: Control(email) + HL(phone,fax)
	contact := ui.Elements[1]
	if len(contact.Elements) != 2 {
		t.Fatalf("expected 2 elements in Contact, got %d", len(contact.Elements))
	}

	if contact.Elements[0].Type != "Control" {
		t.Errorf("expected Control for email, got %q", contact.Elements[0].Type)
	}

	hlContact := contact.Elements[1]
	if hlContact.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout for phone/fax, got %q", hlContact.Type)
	}
}

func TestGenerateUISchema_HorizontalLayout_AllFields(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalAllFields{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All 3 fields horizontal → single HorizontalLayout.
	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	hl := ui.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout, got %q", hl.Type)
	}

	if len(hl.Elements) != 3 {
		t.Errorf("expected 3 controls in HL, got %d", len(hl.Elements))
	}
}

func TestGenerateUISchema_HorizontalLayout_JSON(t *testing.T) {
	ui, err := parser.GenerateUISchema(HorizontalPair{})
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

	elements := raw["elements"].([]any)
	hl := elements[0].(map[string]any)

	if hl["type"] != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout in JSON, got %v", hl["type"])
	}

	hlElements := hl["elements"].([]any)
	if len(hlElements) != 2 {
		t.Errorf("expected 2 elements in HorizontalLayout JSON, got %d", len(hlElements))
	}
}

// --- Named Layout Groups ---

func TestGenerateUISchema_NamedLayoutGroup_NonAdjacent(t *testing.T) {
	type Form struct {
		City    string `json:"city" form:"layout=horizontal:addr"`
		Email   string `json:"email"`
		ZipCode string `json:"zip_code" form:"layout=horizontal:addr"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be: HorizontalLayout(city, zip_code), Control(email)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	hl := ui.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Fatalf("expected 2 elements in HorizontalLayout, got %d", len(hl.Elements))
	}

	if hl.Elements[0].Scope != "#/properties/city" {
		t.Errorf("expected city scope, got %q", hl.Elements[0].Scope)
	}

	if hl.Elements[1].Scope != "#/properties/zip_code" {
		t.Errorf("expected zip_code scope, got %q", hl.Elements[1].Scope)
	}

	// email should be a plain control
	if ui.Elements[1].Type != "Control" {
		t.Errorf("expected Control, got %q", ui.Elements[1].Type)
	}
}

func TestGenerateUISchema_NamedLayoutGroup_DifferentGroups(t *testing.T) {
	type Form struct {
		City    string `json:"city" form:"layout=horizontal:addr"`
		Phone   string `json:"phone" form:"layout=horizontal:contact"`
		ZipCode string `json:"zip_code" form:"layout=horizontal:addr"`
		Fax     string `json:"fax" form:"layout=horizontal:contact"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be: HorizontalLayout(city, zip_code), HorizontalLayout(phone, fax)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	addr := ui.Elements[0]
	if addr.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout for addr, got %q", addr.Type)
	}

	if len(addr.Elements) != 2 {
		t.Fatalf("expected 2 in addr group, got %d", len(addr.Elements))
	}

	if addr.Elements[0].Scope != "#/properties/city" {
		t.Errorf("expected city, got %q", addr.Elements[0].Scope)
	}

	contact := ui.Elements[1]
	if contact.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout for contact, got %q", contact.Type)
	}

	if len(contact.Elements) != 2 {
		t.Fatalf("expected 2 in contact group, got %d", len(contact.Elements))
	}

	if contact.Elements[0].Scope != "#/properties/phone" {
		t.Errorf("expected phone, got %q", contact.Elements[0].Scope)
	}

	if contact.Elements[1].Scope != "#/properties/fax" {
		t.Errorf("expected fax, got %q", contact.Elements[1].Scope)
	}
}

func TestGenerateUISchema_NamedLayoutGroup_UnnamedCompatibility(t *testing.T) {
	// Unnamed layout=horizontal should still group consecutively.
	ui, err := parser.GenerateUISchema(HorizontalPair{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	if ui.Elements[0].Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout, got %q", ui.Elements[0].Type)
	}

	if len(ui.Elements[0].Elements) != 2 {
		t.Errorf("expected 2 in HorizontalLayout, got %d", len(ui.Elements[0].Elements))
	}
}

func TestGenerateUISchema_NamedLayoutGroup_SingleElement(t *testing.T) {
	type Form struct {
		City  string `json:"city" form:"layout=horizontal:addr"`
		Email string `json:"email"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Single named element should not create HorizontalLayout.
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	if ui.Elements[0].Type != "Control" {
		t.Errorf("expected Control for single named, got %q", ui.Elements[0].Type)
	}
}

func TestGenerateUISchema_NamedLayoutGroup_WithCategory(t *testing.T) {
	type Form struct {
		Name    string `json:"name" form:"category=General"`
		City    string `json:"city" form:"category=General;layout=horizontal:addr"`
		Email   string `json:"email" form:"category=General"`
		ZipCode string `json:"zip_code" form:"category=General;layout=horizontal:addr"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != "Categorization" {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	general := ui.Elements[0]
	// Should be: Control(name), HorizontalLayout(city, zip_code), Control(email)
	if len(general.Elements) != 3 {
		t.Fatalf("expected 3 elements in General, got %d", len(general.Elements))
	}

	if general.Elements[0].Type != "Control" {
		t.Errorf("expected Control first, got %q", general.Elements[0].Type)
	}

	hl := general.Elements[1]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Errorf("expected 2 in HorizontalLayout, got %d", len(hl.Elements))
	}

	if general.Elements[2].Type != "Control" {
		t.Errorf("expected Control last, got %q", general.Elements[2].Type)
	}
}

func TestGenerateUISchema_NamedLayoutGroup_InNestedGroup(t *testing.T) {
	type Inner struct {
		Street  string `json:"street" form:"layout=horizontal:loc"`
		Number  int    `json:"number"`
		ZipCode string `json:"zip_code" form:"layout=horizontal:loc"`
	}

	type Form struct {
		Name    string `json:"name"`
		Address Inner  `json:"address"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := ui.Elements[1]
	if group.Type != "Group" {
		t.Fatalf("expected Group, got %q", group.Type)
	}

	// Inside group: HorizontalLayout(street, zip_code), Control(number)
	if len(group.Elements) != 2 {
		t.Fatalf("expected 2 elements in group, got %d", len(group.Elements))
	}

	hl := group.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout in group, got %q", hl.Type)
	}

	if len(hl.Elements) != 2 {
		t.Errorf("expected 2 in HorizontalLayout, got %d", len(hl.Elements))
	}
}

func TestGenerateUISchema_NamedLayoutGroup_MixedNamedUnnamed(t *testing.T) {
	type Form struct {
		A string `json:"a" form:"layout=horizontal:grp"`
		B string `json:"b" form:"layout=horizontal"`
		C string `json:"c" form:"layout=horizontal"`
		D string `json:"d" form:"layout=horizontal:grp"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: HorizontalLayout(a, d), HorizontalLayout(b, c)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	grp := ui.Elements[0]
	if grp.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout for named group, got %q", grp.Type)
	}

	if len(grp.Elements) != 2 {
		t.Fatalf("expected 2 in named group, got %d", len(grp.Elements))
	}

	if grp.Elements[0].Scope != "#/properties/a" {
		t.Errorf("expected a, got %q", grp.Elements[0].Scope)
	}

	if grp.Elements[1].Scope != "#/properties/d" {
		t.Errorf("expected d, got %q", grp.Elements[1].Scope)
	}

	unnamed := ui.Elements[1]
	if unnamed.Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout for unnamed, got %q", unnamed.Type)
	}

	if len(unnamed.Elements) != 2 {
		t.Fatalf("expected 2 in unnamed group, got %d", len(unnamed.Elements))
	}
}

func TestGenerateUISchema_NamedLayoutGroup_OptionsCleanup(t *testing.T) {
	type Form struct {
		City    string `json:"city" form:"layout=horizontal:addr"`
		ZipCode string `json:"zip_code" form:"layout=horizontal:addr"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hl := ui.Elements[0]
	if hl.Type != "HorizontalLayout" {
		t.Fatalf("expected HorizontalLayout, got %q", hl.Type)
	}

	// layout and layoutGroup options should be consumed.
	for _, el := range hl.Elements {
		if el.Options != nil {
			if _, has := el.Options["layout"]; has {
				t.Error("layout option should be consumed")
			}

			if _, has := el.Options["layoutGroup"]; has {
				t.Error("layoutGroup option should be consumed")
			}
		}
	}
}

func TestGenerateUISchema_NamedLayoutGroup_JSON(t *testing.T) {
	type Form struct {
		City    string `json:"city" form:"layout=horizontal:addr"`
		Email   string `json:"email"`
		ZipCode string `json:"zip_code" form:"layout=horizontal:addr"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	elements := parsed["elements"].([]any)
	hl := elements[0].(map[string]any)

	if hl["type"] != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout in JSON, got %v", hl["type"])
	}

	hlElements := hl["elements"].([]any)
	if len(hlElements) != 2 {
		t.Errorf("expected 2 in HorizontalLayout, got %d", len(hlElements))
	}

	// Verify no layout/layoutGroup options leak into JSON.
	for _, raw := range hlElements {
		ctrl := raw.(map[string]any)
		if opts, ok := ctrl["options"]; ok {
			optsMap := opts.(map[string]any)
			if _, has := optsMap["layout"]; has {
				t.Error("layout should not appear in JSON")
			}

			if _, has := optsMap["layoutGroup"]; has {
				t.Error("layoutGroup should not appear in JSON")
			}
		}
	}
}

// --- Array Detail (slice of structs) ---

type ArrayDetailItem struct {
	Numbers int    `json:"numbers"`
	Bonus   int    `json:"bonus"`
	Label   string `json:"label" form:"label=Set Label"`
}

func TestGenerateUISchema_ArrayDetail_SliceOfStructs(t *testing.T) {
	type Form struct {
		Name  string            `json:"name"`
		Items []ArrayDetailItem `json:"items" form:"label=Items"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	ctrl := ui.Elements[1]
	if ctrl.Type != "Control" {
		t.Fatalf("expected Control, got %q", ctrl.Type)
	}

	if ctrl.Scope != "#/properties/items" {
		t.Errorf("expected scope #/properties/items, got %q", ctrl.Scope)
	}

	if ctrl.Label != "Items" {
		t.Errorf("expected label 'Items', got %q", ctrl.Label)
	}

	if ctrl.Options == nil {
		t.Fatal("expected options with detail")
	}

	detail, ok := ctrl.Options["detail"].(*schema.UISchemaElement)
	if !ok {
		t.Fatalf("expected detail to be *UISchemaElement, got %T", ctrl.Options["detail"])
	}

	if detail.Type != "VerticalLayout" {
		t.Errorf("expected VerticalLayout in detail, got %q", detail.Type)
	}

	if len(detail.Elements) != 3 {
		t.Fatalf("expected 3 elements in detail, got %d", len(detail.Elements))
	}

	if detail.Elements[0].Scope != "#/properties/numbers" {
		t.Errorf("expected #/properties/numbers, got %q", detail.Elements[0].Scope)
	}

	if detail.Elements[2].Label != "Set Label" {
		t.Errorf("expected label 'Set Label', got %q", detail.Elements[2].Label)
	}
}

func TestGenerateUISchema_ArrayDetail_SliceOfPtrStructs(t *testing.T) {
	type Form struct {
		Items []*ArrayDetailItem `json:"items"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrl := ui.Elements[0]
	if ctrl.Options == nil {
		t.Fatal("expected options with detail")
	}

	detail, ok := ctrl.Options["detail"].(*schema.UISchemaElement)
	if !ok {
		t.Fatalf("expected detail to be *UISchemaElement, got %T", ctrl.Options["detail"])
	}

	if len(detail.Elements) != 3 {
		t.Errorf("expected 3 elements in detail, got %d", len(detail.Elements))
	}
}

func TestGenerateUISchema_ArrayDetail_PrimitiveSlice_NoDetail(t *testing.T) {
	type Form struct {
		Tags []string `json:"tags"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrl := ui.Elements[0]
	if ctrl.Type != "Control" {
		t.Errorf("expected Control, got %q", ctrl.Type)
	}

	// Primitive slices should NOT have options.detail.
	if ctrl.Options != nil {
		if _, has := ctrl.Options["detail"]; has {
			t.Error("primitive slice should not have options.detail")
		}
	}
}

func TestGenerateUISchema_ArrayDetail_WithCategory(t *testing.T) {
	type Form struct {
		Name  string            `json:"name" form:"category=General"`
		Items []ArrayDetailItem `json:"items" form:"label=Items;category=Data"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != "Categorization" {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	// Find Data category.
	var dataCat *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.Label == "Data" {
			dataCat = el
			break
		}
	}

	if dataCat == nil {
		t.Fatal("Data category not found")
	}

	ctrl := dataCat.Elements[0]
	if ctrl.Options == nil {
		t.Fatal("expected options with detail in category")
	}

	if _, ok := ctrl.Options["detail"].(*schema.UISchemaElement); !ok {
		t.Errorf("expected detail in category item")
	}
}

func TestGenerateUISchema_ArrayDetail_HorizontalInDetail(t *testing.T) {
	type DetailItem struct {
		City    string `json:"city" form:"layout=horizontal"`
		ZipCode string `json:"zip_code" form:"layout=horizontal"`
		Note    string `json:"note"`
	}

	type Form struct {
		Entries []DetailItem `json:"entries"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrl := ui.Elements[0]
	detail, ok := ctrl.Options["detail"].(*schema.UISchemaElement)
	if !ok {
		t.Fatal("expected detail")
	}

	// Should be: HorizontalLayout(city, zip_code), Control(note)
	if len(detail.Elements) != 2 {
		t.Fatalf("expected 2 elements in detail (HL + Control), got %d", len(detail.Elements))
	}

	if detail.Elements[0].Type != "HorizontalLayout" {
		t.Errorf("expected HorizontalLayout in detail, got %q", detail.Elements[0].Type)
	}

	if len(detail.Elements[0].Elements) != 2 {
		t.Errorf("expected 2 in HorizontalLayout, got %d", len(detail.Elements[0].Elements))
	}
}

func TestGenerateUISchema_ArrayDetail_NestedStructInItem(t *testing.T) {
	type Inner struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Item struct {
		Name    string `json:"name"`
		Address Inner  `json:"address"`
	}

	type Form struct {
		Items []Item `json:"items"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrl := ui.Elements[0]
	detail, ok := ctrl.Options["detail"].(*schema.UISchemaElement)
	if !ok {
		t.Fatal("expected detail")
	}

	// Should be: Control(name), Group(Address)
	if len(detail.Elements) != 2 {
		t.Fatalf("expected 2 elements in detail, got %d", len(detail.Elements))
	}

	if detail.Elements[0].Type != "Control" {
		t.Errorf("expected Control, got %q", detail.Elements[0].Type)
	}

	group := detail.Elements[1]
	if group.Type != "Group" {
		t.Errorf("expected Group, got %q", group.Type)
	}

	if len(group.Elements) != 2 {
		t.Errorf("expected 2 elements in group, got %d", len(group.Elements))
	}
}

func TestGenerateUISchema_ArrayDetail_JSON(t *testing.T) {
	type Form struct {
		Items []ArrayDetailItem `json:"items" form:"label=Items"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	elements := parsed["elements"].([]any)
	ctrl := elements[0].(map[string]any)

	opts, ok := ctrl["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options in JSON")
	}

	detail, ok := opts["detail"].(map[string]any)
	if !ok {
		t.Fatal("expected detail object in JSON")
	}

	if detail["type"] != "VerticalLayout" {
		t.Errorf("expected VerticalLayout in JSON detail, got %v", detail["type"])
	}

	detailElements := detail["elements"].([]any)
	if len(detailElements) != 3 {
		t.Errorf("expected 3 elements in JSON detail, got %d", len(detailElements))
	}

	// Verify scopes are relative (#/properties/...)
	first := detailElements[0].(map[string]any)
	if first["scope"] != "#/properties/numbers" {
		t.Errorf("expected relative scope, got %v", first["scope"])
	}
}

func TestGenerateUISchema_ArrayDetail_EmptyStruct_NoDetail(t *testing.T) {
	type Empty struct{}

	type Form struct {
		Items []Empty `json:"items"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrl := ui.Elements[0]
	// Empty struct → no detail (nil or no detail key).
	if ctrl.Options != nil {
		if _, has := ctrl.Options["detail"]; has {
			t.Error("empty struct should not produce detail")
		}
	}
}

// --- Stage 11.1: Rules on Category ---

type CategoryRuleShow struct {
	Provide   bool   `json:"provideAddress"`
	FirstName string `json:"firstName" form:"category=Personal"`
	Street    string `json:"street" form:"category=Address;visibleIf=provideAddress:true"`
	City      string `json:"city" form:"category=Address"`
}

func TestGenerateUISchema_CategoryRule_Show(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleShow{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization root, got %q", ui.Type)
	}

	// Find the Address category.
	var address *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.Label == "Address" {
			address = el
			break
		}
	}

	if address == nil {
		t.Fatal("expected Address category")
	}

	if address.Rule == nil {
		t.Fatal("expected rule on Address category")
	}

	if address.Rule.Effect != schema.EffectShow {
		t.Errorf("expected SHOW effect, got %q", address.Rule.Effect)
	}

	if address.Rule.Condition.Scope != "#/properties/provideAddress" {
		t.Errorf("expected scope '#/properties/provideAddress', got %q", address.Rule.Condition.Scope)
	}

	if address.Rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", address.Rule.Condition.Schema.Const)
	}

	// Controls inside Address should NOT have the rule hint options.
	for _, el := range address.Elements {
		if el.Options != nil {
			if _, ok := el.Options["categoryRuleEffect"]; ok {
				t.Error("categoryRuleEffect should be consumed")
			}

			if _, ok := el.Options["categoryRuleExpr"]; ok {
				t.Error("categoryRuleExpr should be consumed")
			}
		}
	}
}

type CategoryRuleHide struct {
	Role   string `json:"role"`
	Public string `json:"public" form:"category=Public"`
	Secret string `json:"secret" form:"category=Admin;hideIf=role:guest"`
	Config string `json:"config" form:"category=Admin"`
}

func TestGenerateUISchema_CategoryRule_Hide(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleHide{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var admin *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.Label == "Admin" {
			admin = el
			break
		}
	}

	if admin == nil {
		t.Fatal("expected Admin category")
	}

	if admin.Rule == nil {
		t.Fatal("expected rule on Admin category")
	}

	if admin.Rule.Effect != schema.EffectHide {
		t.Errorf("expected HIDE effect, got %q", admin.Rule.Effect)
	}

	if admin.Rule.Condition.Schema.Const != "guest" {
		t.Errorf("expected const 'guest', got %v", admin.Rule.Condition.Schema.Const)
	}
}

func TestGenerateUISchema_CategoryRule_NoRegression(t *testing.T) {
	// Existing CategorizedForm should still work without rules.
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategorizedForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	for _, cat := range ui.Elements {
		if cat.Rule != nil {
			t.Errorf("category %q should have no rule, got %+v", cat.Label, cat.Rule)
		}
	}
}

type CategoryRuleMixed struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name" form:"category=Basic"`
	Age     int    `json:"age" form:"category=Basic"`
	Extra   string `json:"extra" form:"category=Advanced;visibleIf=enabled:true"`
}

func TestGenerateUISchema_CategoryRule_MixedCategories(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleMixed{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	var basic, advanced *schema.UISchemaElement
	for _, el := range ui.Elements {
		switch el.Label {
		case "Basic":
			basic = el
		case "Advanced":
			advanced = el
		}
	}

	if basic == nil {
		t.Fatal("expected Basic category")
	}

	if basic.Rule != nil {
		t.Error("Basic category should have no rule")
	}

	if advanced == nil {
		t.Fatal("expected Advanced category")
	}

	if advanced.Rule == nil {
		t.Fatal("expected rule on Advanced category")
	}

	if advanced.Rule.Effect != schema.EffectShow {
		t.Errorf("expected SHOW, got %q", advanced.Rule.Effect)
	}
}

type CategoryRuleEnable struct {
	Active  bool   `json:"active"`
	Setting string `json:"setting" form:"category=Settings;enableIf=active:true"`
}

func TestGenerateUISchema_CategoryRule_Enable(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleEnable{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var settings *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.Label == "Settings" {
			settings = el
			break
		}
	}

	if settings == nil {
		t.Fatal("expected Settings category")
	}

	if settings.Rule == nil {
		t.Fatal("expected rule on Settings category")
	}

	if settings.Rule.Effect != schema.EffectEnable {
		t.Errorf("expected ENABLE effect, got %q", settings.Rule.Effect)
	}
}

type CategoryRuleDisable struct {
	Locked bool   `json:"locked"`
	Data   string `json:"data" form:"category=Edit;disableIf=locked:true"`
}

func TestGenerateUISchema_CategoryRule_Disable(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleDisable{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var edit *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.Label == "Edit" {
			edit = el
			break
		}
	}

	if edit == nil {
		t.Fatal("expected Edit category")
	}

	if edit.Rule == nil {
		t.Fatal("expected rule on Edit category")
	}

	if edit.Rule.Effect != schema.EffectDisable {
		t.Errorf("expected DISABLE effect, got %q", edit.Rule.Effect)
	}
}

func TestGenerateUISchema_CategoryRule_JSON(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryRuleShow{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	// Check the Address category has a rule in JSON.
	elements := parsed["elements"].([]any)

	var addressCat map[string]any
	for _, el := range elements {
		cat := el.(map[string]any)
		if cat["label"] == "Address" {
			addressCat = cat
			break
		}
	}

	if addressCat == nil {
		t.Fatal("expected Address category in JSON")
	}

	ruleObj, ok := addressCat["rule"].(map[string]any)
	if !ok {
		t.Fatal("expected rule object on Address category")
	}

	if ruleObj["effect"] != "SHOW" {
		t.Errorf("expected SHOW effect in JSON, got %v", ruleObj["effect"])
	}
}

// --- Stage 11.2: i18n on Category ---

type CategoryI18n struct {
	Name  string `json:"name" form:"category=Personal;i18n=category.personal"`
	Email string `json:"email" form:"category=Personal"`
	Role  string `json:"role" form:"category=Work;i18n=category.work"`
}

func TestGenerateUISchema_CategoryI18n_Key(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryI18n{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeCategorization {
		t.Fatalf("expected Categorization root, got %q", ui.Type)
	}

	var personal, work *schema.UISchemaElement
	for _, el := range ui.Elements {
		switch el.Label {
		case "Personal":
			personal = el
		case "Work":
			work = el
		}
	}

	if personal == nil {
		t.Fatal("expected Personal category")
	}

	if personal.I18n != "category.personal" {
		t.Errorf("expected i18n 'category.personal', got %q", personal.I18n)
	}

	if work == nil {
		t.Fatal("expected Work category")
	}

	if work.I18n != "category.work" {
		t.Errorf("expected i18n 'category.work', got %q", work.I18n)
	}

	// i18n hint should be consumed from child options.
	for _, cat := range ui.Elements {
		for _, el := range cat.Elements {
			if el.Options != nil {
				if _, ok := el.Options["categoryI18n"]; ok {
					t.Error("categoryI18n hint should be consumed")
				}
			}
		}
	}
}

func TestGenerateUISchema_CategoryI18n_NoI18n(t *testing.T) {
	// Existing CategorizedForm has no i18n — should work without regression.
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategorizedForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, cat := range ui.Elements {
		if cat.I18n != "" {
			t.Errorf("category %q should have no i18n, got %q", cat.Label, cat.I18n)
		}
	}
}

func TestGenerateUISchema_CategoryI18n_Translator(t *testing.T) {
	tr := schema.NewMapTranslator(map[string]map[string]string{
		"uk": {
			"category.personal": "Особисте",
			"category.work":     "Робота",
		},
	})

	opts := schema.Options{
		Draft:      "draft-07",
		Translator: tr,
		Locale:     "uk",
	}

	ui, err := parser.GenerateUISchemaWithOptions(CategoryI18n{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var personal, work *schema.UISchemaElement
	for _, el := range ui.Elements {
		switch el.I18n {
		case "category.personal":
			personal = el
		case "category.work":
			work = el
		}
	}

	if personal == nil {
		t.Fatal("expected Personal category")
	}

	if personal.Label != "Особисте" {
		t.Errorf("expected translated label 'Особисте', got %q", personal.Label)
	}

	if work == nil {
		t.Fatal("expected Work category")
	}

	if work.Label != "Робота" {
		t.Errorf("expected translated label 'Робота', got %q", work.Label)
	}
}

func TestGenerateUISchema_CategoryI18n_JSON(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryI18n{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	elements := parsed["elements"].([]any)

	var personalCat map[string]any
	for _, el := range elements {
		cat := el.(map[string]any)
		if cat["label"] == "Personal" {
			personalCat = cat
			break
		}
	}

	if personalCat == nil {
		t.Fatal("expected Personal category in JSON")
	}

	if personalCat["i18n"] != "category.personal" {
		t.Errorf("expected i18n 'category.personal' in JSON, got %v", personalCat["i18n"])
	}
}

type CategoryI18nWithRule struct {
	Active bool   `json:"active"`
	Name   string `json:"name" form:"category=Profile;i18n=category.profile;visibleIf=active:true"`
}

func TestGenerateUISchema_CategoryI18n_WithRule(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(CategoryI18nWithRule{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profile *schema.UISchemaElement
	for _, el := range ui.Elements {
		if el.I18n == "category.profile" {
			profile = el
			break
		}
	}

	if profile == nil {
		t.Fatal("expected Profile category")
	}

	if profile.I18n != "category.profile" {
		t.Errorf("expected i18n 'category.profile', got %q", profile.I18n)
	}

	if profile.Rule == nil {
		t.Fatal("expected rule on Profile category")
	}

	if profile.Rule.Effect != schema.EffectShow {
		t.Errorf("expected SHOW effect, got %q", profile.Rule.Effect)
	}
}

// --- Stage 11.3: Rules on nested structs (Group) ---

type GroupRuleAddress struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type GroupRuleForm struct {
	ProvideAddress bool             `json:"provideAddress"`
	Name           string           `json:"name"`
	Address        GroupRuleAddress `json:"address" visibleIf:"provideAddress=true"`
}

func TestGenerateUISchema_GroupRule_VisibleIf(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupRuleForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Root is VerticalLayout with: Control(provideAddress), Control(name), Group(Address).
	if ui.Type != "VerticalLayout" {
		t.Fatalf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(ui.Elements))
	}

	group := ui.Elements[2]
	if group.Type != "Group" {
		t.Fatalf("expected Group, got %q", group.Type)
	}

	if group.Rule == nil {
		t.Fatal("expected rule on Group")
	}

	if group.Rule.Effect != schema.EffectShow {
		t.Errorf("expected SHOW effect, got %q", group.Rule.Effect)
	}

	if group.Rule.Condition.Scope != "#/properties/provideAddress" {
		t.Errorf("expected scope '#/properties/provideAddress', got %q", group.Rule.Condition.Scope)
	}

	if group.Rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", group.Rule.Condition.Schema.Const)
	}
}

type GroupRuleHideForm struct {
	IsAdmin bool             `json:"isAdmin"`
	Secret  GroupRuleAddress `json:"secret" hideIf:"isAdmin=false"`
}

func TestGenerateUISchema_GroupRule_HideIf(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupRuleHideForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := ui.Elements[1]
	if group.Type != "Group" {
		t.Fatalf("expected Group, got %q", group.Type)
	}

	if group.Rule == nil {
		t.Fatal("expected rule on Group")
	}

	if group.Rule.Effect != schema.EffectHide {
		t.Errorf("expected HIDE effect, got %q", group.Rule.Effect)
	}

	if group.Rule.Condition.Schema.Const != false {
		t.Errorf("expected const false, got %v", group.Rule.Condition.Schema.Const)
	}
}

type GroupRuleEnableForm struct {
	Active  bool             `json:"active"`
	Details GroupRuleAddress `json:"details" enableIf:"active=true"`
}

func TestGenerateUISchema_GroupRule_EnableIf(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupRuleEnableForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := ui.Elements[1]

	if group.Rule == nil {
		t.Fatal("expected rule on Group")
	}

	if group.Rule.Effect != schema.EffectEnable {
		t.Errorf("expected ENABLE effect, got %q", group.Rule.Effect)
	}
}

type GroupRuleDisableForm struct {
	Locked bool             `json:"locked"`
	Config GroupRuleAddress `json:"config" disableIf:"locked=true"`
}

func TestGenerateUISchema_GroupRule_DisableIf(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupRuleDisableForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := ui.Elements[1]

	if group.Rule == nil {
		t.Fatal("expected rule on Group")
	}

	if group.Rule.Effect != schema.EffectDisable {
		t.Errorf("expected DISABLE effect, got %q", group.Rule.Effect)
	}
}

type GroupNoRuleForm struct {
	Name    string           `json:"name"`
	Address GroupRuleAddress `json:"address"`
}

func TestGenerateUISchema_GroupRule_NoRegression(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupNoRuleForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	group := ui.Elements[1]
	if group.Type != "Group" {
		t.Fatalf("expected Group, got %q", group.Type)
	}

	if group.Rule != nil {
		t.Errorf("expected no rule on Group without rule tags, got %+v", group.Rule)
	}
}

func TestGenerateUISchema_GroupRule_JSON(t *testing.T) {
	opts := schema.DefaultOptions()

	ui, err := parser.GenerateUISchemaWithOptions(GroupRuleForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	elements := parsed["elements"].([]any)
	groupObj := elements[2].(map[string]any)

	if groupObj["type"] != "Group" {
		t.Fatalf("expected Group in JSON, got %v", groupObj["type"])
	}

	ruleObj, ok := groupObj["rule"].(map[string]any)
	if !ok {
		t.Fatal("expected rule on Group in JSON")
	}

	if ruleObj["effect"] != "SHOW" {
		t.Errorf("expected SHOW in JSON, got %v", ruleObj["effect"])
	}
}

// --- Group in Category ---

type NestedGroupDrawSetup struct {
	Numbers int `json:"numbers"`
	Bonus   int `json:"bonus"`
}

type GroupCategoryForm struct {
	GameName  string               `json:"game_name" form:"label=Game name;category=General"`
	DrawSetup NestedGroupDrawSetup `json:"draw_setup" form:"label=Draw setup;category=General"`
	Logic     string               `json:"logic" form:"label=Logic;category=Logic"`
}

func TestGenerateUISchema_GroupInCategory(t *testing.T) {
	ui, err := parser.GenerateUISchema(GroupCategoryForm{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != "Categorization" {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	// Should have 2 categories: General and Logic (no "Other").
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(ui.Elements))
	}

	general := ui.Elements[0]
	if general.Label != "General" {
		t.Errorf("expected first category 'General', got %q", general.Label)
	}

	// General should have: Control(game_name) + Group(draw_setup).
	if len(general.Elements) != 2 {
		t.Fatalf("expected 2 elements in General, got %d", len(general.Elements))
	}

	if general.Elements[0].Type != "Control" {
		t.Errorf("expected first element Control, got %q", general.Elements[0].Type)
	}

	group := general.Elements[1]
	if group.Type != "Group" {
		t.Errorf("expected second element Group, got %q", group.Type)
	}

	if group.Label != "Draw setup" {
		t.Errorf("expected Group label 'Draw setup', got %q", group.Label)
	}

	// Group should not leak 'category' option.
	if group.Options != nil {
		if _, has := group.Options["category"]; has {
			t.Error("category option should be consumed, not present on Group")
		}
	}

	logic := ui.Elements[1]
	if logic.Label != "Logic" {
		t.Errorf("expected second category 'Logic', got %q", logic.Label)
	}
}

func TestGenerateUISchema_GroupInCategory_NoCategory_FallsToOther(t *testing.T) {
	type Form struct {
		Name    string               `json:"name" form:"category=General"`
		Details NestedGroupDrawSetup `json:"details" form:"label=Details"`
	}

	ui, err := parser.GenerateUISchema(Form{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != "Categorization" {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	// Should have 2 categories: General and Other.
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(ui.Elements))
	}

	other := ui.Elements[1]
	if other.Label != "Other" {
		t.Errorf("expected 'Other', got %q", other.Label)
	}

	if other.Elements[0].Type != "Group" {
		t.Errorf("expected Group in Other, got %q", other.Elements[0].Type)
	}
}

type GroupCategoryRuleForm struct {
	Provide   bool                 `json:"provide" form:"category=General"`
	DrawSetup NestedGroupDrawSetup `json:"draw_setup" form:"label=Draw setup;category=Setup;visibleIf=provide:true"`
}

func TestGenerateUISchema_GroupInCategory_WithRule(t *testing.T) {
	ui, err := parser.GenerateUISchema(GroupCategoryRuleForm{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != "Categorization" {
		t.Fatalf("expected Categorization, got %q", ui.Type)
	}

	setup := ui.Elements[1]
	if setup.Label != "Setup" {
		t.Errorf("expected 'Setup', got %q", setup.Label)
	}

	if setup.Rule == nil {
		t.Fatal("expected rule on Setup category")
	}

	if setup.Rule.Effect != schema.EffectShow {
		t.Errorf("expected SHOW, got %q", setup.Rule.Effect)
	}

	if setup.Rule.Condition.Scope != "#/properties/provide" {
		t.Errorf("expected scope '#/properties/provide', got %q", setup.Rule.Condition.Scope)
	}
}

type GroupCategoryI18nForm struct {
	Name    string               `json:"name" form:"category=Personal;i18n=cat.personal"`
	Details NestedGroupDrawSetup `json:"details" form:"label=Draw;category=Setup;i18n=cat.setup"`
}

func TestGenerateUISchema_GroupInCategory_WithI18n(t *testing.T) {
	opts := schema.Options{
		Translator: schema.NewMapTranslator(map[string]map[string]string{
			"en": {
				"cat.personal": "My Personal",
				"cat.setup":    "My Setup",
			},
		}),
		Locale: "en",
	}

	ui, err := parser.GenerateUISchemaWithOptions(GroupCategoryI18nForm{}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	setup := ui.Elements[1]
	if setup.I18n != "cat.setup" {
		t.Errorf("expected i18n 'cat.setup', got %q", setup.I18n)
	}

	if setup.Label != "My Setup" {
		t.Errorf("expected translated label 'My Setup', got %q", setup.Label)
	}
}

func TestGenerateUISchema_GroupInCategory_JSON(t *testing.T) {
	ui, err := parser.GenerateUISchema(GroupCategoryForm{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.MarshalIndent(ui, "", "  ")
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	cats := parsed["elements"].([]any)
	general := cats[0].(map[string]any)

	elems := general["elements"].([]any)
	groupObj := elems[1].(map[string]any)

	if groupObj["type"] != "Group" {
		t.Errorf("expected Group in JSON, got %v", groupObj["type"])
	}

	if groupObj["label"] != "Draw setup" {
		t.Errorf("expected 'Draw setup' in JSON, got %v", groupObj["label"])
	}

	// Should have nested controls.
	groupElems := groupObj["elements"].([]any)
	if len(groupElems) != 2 {
		t.Errorf("expected 2 elements in Group, got %d", len(groupElems))
	}
}
