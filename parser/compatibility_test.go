package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/holdemlab/ui-json-schema/parser"
)

// JSONFormsUser is a realistic struct that exercises all supported features.
type JSONFormsUser struct {
	ID        int    `json:"id" form:"hidden"`
	Name      string `json:"name" required:"true" form:"label=Full name"`
	Email     string `json:"email" required:"true" format:"email"`
	IsActive  bool   `json:"is_active" default:"true"`
	Role      string `json:"role" enum:"admin,user,moderator"`
	Bio       string `json:"bio" form:"multiline"`
	CreatedAt string `json:"created_at" form:"readonly"`
	Details   string `json:"details" visibleIf:"is_active=true"`
}

// TestJSONFormsCompatibility_StructSchema verifies that the generated JSON Schema
// follows the JSON Forms / JSON Schema Draft 7 contract.
func TestJSONFormsCompatibility_StructSchema(t *testing.T) {
	s, err := parser.GenerateJSONSchema(JSONFormsUser{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// $schema must be Draft 7.
	if raw["$schema"] != schemaDraft7 {
		t.Errorf("expected Draft 7 $schema, got %v", raw["$schema"])
	}

	// Root type must be "object".
	if raw["type"] != typeObject {
		t.Errorf("expected root type 'object', got %v", raw["type"])
	}

	// Must have properties.
	props, ok := raw["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing properties object")
	}

	// Verify field types.
	typeChecks := map[string]string{
		"id":         "integer",
		"name":       "string",
		"email":      "string",
		"is_active":  "boolean",
		"role":       "string",
		"bio":        "string",
		"created_at": "string",
		"details":    "string",
	}

	for field, expectedType := range typeChecks {
		prop, ok := props[field].(map[string]any)
		if !ok {
			t.Errorf("missing property %q", field)
			continue
		}

		if prop["type"] != expectedType {
			t.Errorf("property %q: expected type %q, got %v", field, expectedType, prop["type"])
		}
	}

	// Verify required array.
	required, ok := raw["required"].([]any)
	if !ok {
		t.Fatal("missing required array")
	}

	reqSet := make(map[string]bool)
	for _, r := range required {
		reqSet[r.(string)] = true
	}

	if !reqSet["name"] || !reqSet["email"] {
		t.Errorf("expected name and email in required, got %v", required)
	}

	// Verify format on email.
	emailProp := props["email"].(map[string]any)
	if emailProp["format"] != "email" {
		t.Errorf("expected email format, got %v", emailProp["format"])
	}

	// Verify default on is_active.
	activeProp := props["is_active"].(map[string]any)
	if activeProp["default"] != true {
		t.Errorf("expected is_active default true, got %v", activeProp["default"])
	}

	// Verify enum on role.
	roleProp := props["role"].(map[string]any)
	enumVals, ok := roleProp["enum"].([]any)
	if !ok || len(enumVals) != 3 {
		t.Errorf("expected 3 enum values for role, got %v", roleProp["enum"])
	}
}

// TestJSONFormsCompatibility_UISchema verifies that the generated UI Schema
// follows the JSON Forms UI Schema contract.
func TestJSONFormsCompatibility_UISchema(t *testing.T) {
	ui, err := parser.GenerateUISchema(JSONFormsUser{})
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

	// Root must be VerticalLayout.
	if raw["type"] != "VerticalLayout" {
		t.Errorf("expected root type 'VerticalLayout', got %v", raw["type"])
	}

	// Must have elements.
	elements, ok := raw["elements"].([]any)
	if !ok {
		t.Fatal("missing elements array")
	}

	// Hidden field (id) should be excluded.
	for _, el := range elements {
		elem := el.(map[string]any)
		if elem["scope"] == "#/properties/id" {
			t.Error("hidden field 'id' should not appear in UI Schema")
		}
	}

	// Check that we have 7 elements (8 fields - 1 hidden).
	if len(elements) != 7 {
		t.Fatalf("expected 7 elements (8 fields - 1 hidden), got %d", len(elements))
	}

	// Verify Control elements have required fields.
	for _, el := range elements {
		elem := el.(map[string]any)
		if elem["type"] != typeControl {
			t.Errorf("expected all elements to be Control, got %v", elem["type"])
		}

		if elem["scope"] == nil || elem["scope"] == "" {
			t.Error("Control element must have a non-empty scope")
		}
	}

	// Find specific elements and verify their properties.
	elemMap := make(map[string]map[string]any)
	for _, el := range elements {
		elem := el.(map[string]any)
		scope, _ := elem["scope"].(string)
		elemMap[scope] = elem
	}

	// Name should have label.
	nameEl := elemMap["#/properties/name"]
	if nameEl == nil {
		t.Fatal("missing name control")
	}

	if nameEl["label"] != labelFullName {
		t.Errorf("expected label '%s', got %v", labelFullName, nameEl["label"])
	}

	// Bio should have multiline option.
	bioEl := elemMap["#/properties/bio"]
	if bioEl == nil {
		t.Fatal("missing bio control")
	}

	opts, ok := bioEl["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options on bio")
	}

	if opts["multi"] != true {
		t.Errorf("expected multi=true for bio, got %v", opts["multi"])
	}

	// CreatedAt should have readonly option.
	createdEl := elemMap["#/properties/created_at"]
	if createdEl == nil {
		t.Fatal("missing created_at control")
	}

	crOpts, ok := createdEl["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options on created_at")
	}

	if crOpts["readonly"] != true {
		t.Errorf("expected readonly=true for created_at, got %v", crOpts["readonly"])
	}

	// Details should have a SHOW rule.
	detailsEl := elemMap["#/properties/details"]
	if detailsEl == nil {
		t.Fatal("missing details control")
	}

	rule, ok := detailsEl["rule"].(map[string]any)
	if !ok {
		t.Fatal("expected rule on details")
	}

	if rule["effect"] != "SHOW" {
		t.Errorf("expected SHOW effect, got %v", rule["effect"])
	}

	cond, ok := rule["condition"].(map[string]any)
	if !ok {
		t.Fatal("expected condition in rule")
	}

	if cond["scope"] != "#/properties/is_active" {
		t.Errorf("expected condition scope '#/properties/is_active', got %v", cond["scope"])
	}

	condSchema, ok := cond["schema"].(map[string]any)
	if !ok {
		t.Fatal("expected schema in condition")
	}

	if condSchema["const"] != true {
		t.Errorf("expected const true, got %v", condSchema["const"])
	}
}

// TestJSONFormsCompatibility_FromJSON verifies JSON Forms compatibility for JSON-generated schemas.
func TestJSONFormsCompatibility_FromJSON(t *testing.T) {
	input := []byte(`{
		"name": "John",
		"age": 30,
		"is_active": true,
		"score": 4.5,
		"tags": ["go", "json"],
		"address": {
			"city": "Kyiv",
			"street": "Main St"
		}
	}`)

	s, ui, err := parser.GenerateFromJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify JSON Schema.
	sData, _ := json.Marshal(s)

	var sRaw map[string]any
	if err := json.Unmarshal(sData, &sRaw); err != nil {
		t.Fatalf("invalid schema JSON: %v", err)
	}

	if sRaw["$schema"] != schemaDraft7 {
		t.Errorf("expected Draft 7 $schema, got %v", sRaw["$schema"])
	}

	if sRaw["type"] != typeObject {
		t.Errorf("expected root type 'object', got %v", sRaw["type"])
	}

	props, ok := sRaw["properties"].(map[string]any)
	if !ok {
		t.Fatal("missing properties")
	}

	// Check that address is nested object.
	addrProp, ok := props["address"].(map[string]any)
	if !ok {
		t.Fatal("missing address property")
	}

	if addrProp["type"] != typeObject {
		t.Errorf("expected address type 'object', got %v", addrProp["type"])
	}

	// Check array.
	tagsProp, ok := props["tags"].(map[string]any)
	if !ok {
		t.Fatal("missing tags property")
	}

	if tagsProp["type"] != "array" {
		t.Errorf("expected tags type 'array', got %v", tagsProp["type"])
	}

	// Verify UI Schema.
	uiData, _ := json.Marshal(ui)

	var uiRaw map[string]any
	if err := json.Unmarshal(uiData, &uiRaw); err != nil {
		t.Fatalf("invalid UI schema JSON: %v", err)
	}

	if uiRaw["type"] != "VerticalLayout" {
		t.Errorf("expected UI root 'VerticalLayout', got %v", uiRaw["type"])
	}

	elements, ok := uiRaw["elements"].([]any)
	if !ok {
		t.Fatal("missing UI elements")
	}

	if len(elements) == 0 {
		t.Error("expected at least one UI element")
	}
}
