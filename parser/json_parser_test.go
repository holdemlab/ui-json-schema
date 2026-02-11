package parser_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/holdemlab/ui-json-schema/parser"
)

const typeControl = "Control"

// --- GenerateFromJSON: JSON Schema tests ---

func TestGenerateFromJSON_SimpleObject(t *testing.T) {
	data := []byte(`{"name":"John","age":30,"is_active":true}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Type, "object")

	if len(s.Properties) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(s.Properties))
	}

	assertSchemaType(t, s.Properties["name"].Type, "string")
	assertSchemaType(t, s.Properties["age"].Type, "integer")
	assertSchemaType(t, s.Properties["is_active"].Type, "boolean")
}

func TestGenerateFromJSON_NumberTypes(t *testing.T) {
	data := []byte(`{"integer":42,"float":3.14,"negative":-5,"zero":0}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Properties["integer"].Type, "integer")
	assertSchemaType(t, s.Properties["float"].Type, "number")
	assertSchemaType(t, s.Properties["negative"].Type, "integer")
	assertSchemaType(t, s.Properties["zero"].Type, "integer")
}

func TestGenerateFromJSON_NullValue(t *testing.T) {
	data := []byte(`{"nothing":null}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Properties["nothing"].Type, "null")
}

func TestGenerateFromJSON_NestedObject(t *testing.T) {
	data := []byte(`{"address":{"city":"Kyiv","zip":"01001"}}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	addr, ok := s.Properties["address"]
	if !ok {
		t.Fatal("missing property 'address'")
	}

	assertSchemaType(t, addr.Type, "object")

	if len(addr.Properties) != 2 {
		t.Errorf("expected 2 nested properties, got %d", len(addr.Properties))
	}

	assertSchemaType(t, addr.Properties["city"].Type, "string")
	assertSchemaType(t, addr.Properties["zip"].Type, "string")
}

func TestGenerateFromJSON_DeeplyNestedObject(t *testing.T) {
	data := []byte(`{"level1":{"level2":{"value":"deep"}}}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l1 := s.Properties["level1"]
	if l1 == nil {
		t.Fatal("missing level1")
	}
	assertSchemaType(t, l1.Type, "object")

	l2 := l1.Properties["level2"]
	if l2 == nil {
		t.Fatal("missing level2")
	}
	assertSchemaType(t, l2.Type, "object")

	val := l2.Properties["value"]
	if val == nil {
		t.Fatal("missing value")
	}
	assertSchemaType(t, val.Type, "string")
}

func TestGenerateFromJSON_ArrayOfStrings(t *testing.T) {
	data := []byte(`{"tags":["go","json"]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tags := s.Properties["tags"]
	if tags == nil {
		t.Fatal("missing property 'tags'")
	}

	assertSchemaType(t, tags.Type, "array")

	if tags.Items == nil {
		t.Fatal("expected items schema")
	}
	assertSchemaType(t, tags.Items.Type, "string")
}

func TestGenerateFromJSON_ArrayOfIntegers(t *testing.T) {
	data := []byte(`{"scores":[100,200,300]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scores := s.Properties["scores"]
	assertSchemaType(t, scores.Type, "array")
	assertSchemaType(t, scores.Items.Type, "integer")
}

func TestGenerateFromJSON_ArrayOfObjects(t *testing.T) {
	data := []byte(`{"items":[{"name":"a","value":1}]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := s.Properties["items"]
	assertSchemaType(t, items.Type, "array")
	assertSchemaType(t, items.Items.Type, "object")

	if len(items.Items.Properties) != 2 {
		t.Errorf("expected 2 properties in array items, got %d", len(items.Items.Properties))
	}

	assertSchemaType(t, items.Items.Properties["name"].Type, "string")
	assertSchemaType(t, items.Items.Properties["value"].Type, "integer")
}

func TestGenerateFromJSON_EmptyArray(t *testing.T) {
	data := []byte(`{"empty":[]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	empty := s.Properties["empty"]
	assertSchemaType(t, empty.Type, "array")

	if empty.Items == nil {
		t.Fatal("expected items schema even for empty array")
	}
}

func TestGenerateFromJSON_NoRequired(t *testing.T) {
	data := []byte(`{"name":"test","age":25}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Required) != 0 {
		t.Errorf("expected no required fields, got %v", s.Required)
	}
}

func TestGenerateFromJSON_SchemaField(t *testing.T) {
	data := []byte(`{"x":1}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Schema != schemaDraft7 {
		t.Errorf("expected $schema to be set, got %q", s.Schema)
	}
}

func TestGenerateFromJSON_ValidJSON_Output(t *testing.T) {
	data := []byte(`{"name":"John","age":30,"address":{"city":"Kyiv"},"tags":["go"]}`)

	s, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schemaData, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	if !json.Valid(schemaData) {
		t.Error("generated schema is not valid JSON")
	}

	uiData, err := json.Marshal(ui)
	if err != nil {
		t.Fatalf("failed to marshal UI schema: %v", err)
	}
	if !json.Valid(uiData) {
		t.Error("generated UI schema is not valid JSON")
	}
}

// --- GenerateFromJSON: UI Schema tests ---

func TestGenerateFromJSON_UISchema_Simple(t *testing.T) {
	data := []byte(`{"name":"John","age":30}`)

	_, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	for _, el := range ui.Elements {
		if el.Type != typeControl {
			t.Errorf("expected Control, got %q", el.Type)
		}
	}
}

func TestGenerateFromJSON_UISchema_Scopes(t *testing.T) {
	data := []byte(`{"name":"John","age":30}`)

	_, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scopes := make(map[string]bool)
	for _, el := range ui.Elements {
		scopes[el.Scope] = true
	}

	if !scopes["#/properties/name"] {
		t.Error("missing scope #/properties/name")
	}
	if !scopes["#/properties/age"] {
		t.Error("missing scope #/properties/age")
	}
}

func TestGenerateFromJSON_UISchema_NestedGroup(t *testing.T) {
	data := []byte(`{"address":{"city":"Kyiv","street":"Main St"}}`)

	_, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element (group), got %d", len(ui.Elements))
	}

	group := ui.Elements[0]
	if group.Type != typeGroup {
		t.Errorf("expected Group, got %q", group.Type)
	}

	if group.Label != "address" {
		t.Errorf("expected label 'address', got %q", group.Label)
	}

	if len(group.Elements) != 2 {
		t.Errorf("expected 2 children in group, got %d", len(group.Elements))
	}

	// Verify nested scopes.
	nestedScopes := make(map[string]bool)
	for _, el := range group.Elements {
		nestedScopes[el.Scope] = true
	}

	if !nestedScopes["#/properties/address/properties/city"] {
		t.Error("missing nested scope for city")
	}
	if !nestedScopes["#/properties/address/properties/street"] {
		t.Error("missing nested scope for street")
	}
}

func TestGenerateFromJSON_UISchema_DeeplyNested(t *testing.T) {
	data := []byte(`{"info":{"contact":{"phone":"123"}}}`)

	_, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Root -> Group(info) -> Group(contact) -> Control(phone)
	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 top-level element, got %d", len(ui.Elements))
	}

	infoGroup := ui.Elements[0]
	if infoGroup.Type != typeGroup {
		t.Fatalf("expected Group, got %q", infoGroup.Type)
	}

	if len(infoGroup.Elements) != 1 {
		t.Fatalf("expected 1 child in info group, got %d", len(infoGroup.Elements))
	}

	contactGroup := infoGroup.Elements[0]
	if contactGroup.Type != typeGroup {
		t.Fatalf("expected nested Group, got %q", contactGroup.Type)
	}

	if len(contactGroup.Elements) != 1 {
		t.Fatalf("expected 1 child in contact group, got %d", len(contactGroup.Elements))
	}

	phone := contactGroup.Elements[0]
	if phone.Type != typeControl {
		t.Errorf("expected Control, got %q", phone.Type)
	}
	if phone.Scope != "#/properties/info/properties/contact/properties/phone" {
		t.Errorf("unexpected scope %q", phone.Scope)
	}
}

func TestGenerateFromJSON_UISchema_ArrayFieldIsControl(t *testing.T) {
	data := []byte(`{"tags":["go","json"]}`)

	_, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	el := ui.Elements[0]
	if el.Type != typeControl {
		t.Errorf("expected Control for array, got %q", el.Type)
	}
	if el.Scope != "#/properties/tags" {
		t.Errorf("unexpected scope %q", el.Scope)
	}
}

// --- GenerateFromJSON: error cases ---

func TestGenerateFromJSON_InvalidJSON(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGenerateFromJSON_ArrayRoot(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`[1,2,3]`))
	if err == nil {
		t.Fatal("expected error for array root")
	}
}

func TestGenerateFromJSON_StringRoot(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`"hello"`))
	if err == nil {
		t.Fatal("expected error for string root")
	}
}

func TestGenerateFromJSON_NumberRoot(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`42`))
	if err == nil {
		t.Fatal("expected error for number root")
	}
}

func TestGenerateFromJSON_BoolRoot(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`true`))
	if err == nil {
		t.Fatal("expected error for bool root")
	}
}

func TestGenerateFromJSON_NullRoot(t *testing.T) {
	_, _, err := parser.GenerateFromJSON([]byte(`null`))
	if err == nil {
		t.Fatal("expected error for null root")
	}
}

func TestGenerateFromJSON_EmptyObject(t *testing.T) {
	s, ui, err := parser.GenerateFromJSON([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Properties) != 0 {
		t.Errorf("expected 0 properties, got %d", len(s.Properties))
	}

	if len(ui.Elements) != 0 {
		t.Errorf("expected 0 UI elements, got %d", len(ui.Elements))
	}
}

// --- GenerateFromJSON: deterministic output test ---

func TestGenerateFromJSON_DeterministicSchema(t *testing.T) {
	data := []byte(`{"b":"two","a":"one","c":"three"}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Properties should contain all keys regardless of map iteration order.
	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	expected := []string{"a", "b", "c"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d properties, got %d", len(expected), len(keys))
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("property[%d]: expected %q, got %q", i, expected[i], k)
		}
	}
}

// --- GenerateFromJSON: complex realistic JSON ---

func TestGenerateFromJSON_RealisticPayload(t *testing.T) {
	data := []byte(`{
		"user": {
			"name": "Alice",
			"age": 28,
			"verified": true
		},
		"scores": [95, 87, 100],
		"metadata": null,
		"rating": 4.7
	}`)

	s, ui, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check JSON Schema types.
	assertSchemaType(t, s.Properties["user"].Type, "object")
	assertSchemaType(t, s.Properties["scores"].Type, "array")
	assertSchemaType(t, s.Properties["metadata"].Type, "null")
	assertSchemaType(t, s.Properties["rating"].Type, "number")

	// User nested properties.
	user := s.Properties["user"]
	assertSchemaType(t, user.Properties["name"].Type, "string")
	assertSchemaType(t, user.Properties["age"].Type, "integer")
	assertSchemaType(t, user.Properties["verified"].Type, "boolean")

	// UI Schema should have: Group(user) + Control(scores) + Control(metadata) + Control(rating).
	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 4 {
		t.Fatalf("expected 4 top-level UI elements, got %d", len(ui.Elements))
	}

	// Verify overall JSON validity.
	schemaJSON, _ := json.Marshal(s)
	if !json.Valid(schemaJSON) {
		t.Error("schema output is not valid JSON")
	}
	uiJSON, _ := json.Marshal(ui)
	if !json.Valid(uiJSON) {
		t.Error("UI schema output is not valid JSON")
	}
}

func TestGenerateFromJSON_MixedArray(t *testing.T) {
	// First element determines type.
	data := []byte(`{"items":[42,"string",true]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := s.Properties["items"]
	assertSchemaType(t, items.Type, "array")
	// First element is 42 (integer).
	assertSchemaType(t, items.Items.Type, "integer")
}

func TestGenerateFromJSON_ArrayOfBooleans(t *testing.T) {
	data := []byte(`{"flags":[true,false,true]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flags := s.Properties["flags"]
	assertSchemaType(t, flags.Type, "array")
	assertSchemaType(t, flags.Items.Type, "boolean")
}

func TestGenerateFromJSON_ArrayOfNulls(t *testing.T) {
	data := []byte(`{"nils":[null]}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nils := s.Properties["nils"]
	assertSchemaType(t, nils.Type, "array")
	assertSchemaType(t, nils.Items.Type, "null")
}

func TestGenerateFromJSON_FloatThatLooksLikeInt(t *testing.T) {
	// JSON number 1.0 is parsed as float64 by Go, but Trunc(1.0)==1.0 → integer.
	data := []byte(`{"val":1.0}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Properties["val"].Type, "integer")
}

func TestGenerateFromJSON_LargeFloat(t *testing.T) {
	data := []byte(`{"pi":3.141592653589793}`)

	s, _, err := parser.GenerateFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Properties["pi"].Type, "number")
}
