package parser_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/holdemlab/ui-json-schema/parser"
)

const (
	formatEmail        = "email"
	typeGroup          = "Group"
	typeVerticalLayout = "VerticalLayout"
	typeCategorization = "Categorization"
	typeString         = "string"
	effectShow         = "SHOW"
	schemaDraft7       = "http://json-schema.org/draft-07/schema#"
	labelFullName      = "Full name"
	typeObject         = "object"
)

// --- Test structs ---

type SimpleStruct struct {
	Name     string  `json:"name"`
	Age      int     `json:"age"`
	Score    float64 `json:"score"`
	IsActive bool    `json:"is_active"`
}

type NestedAddress struct {
	City   string `json:"city"`
	Street string `json:"street"`
}

type NestedStruct struct {
	Name    string        `json:"name"`
	Address NestedAddress `json:"address"`
}

type SliceStruct struct {
	Tags   []string `json:"tags"`
	Scores []int    `json:"scores"`
}

type MapStruct struct {
	Metadata map[string]string `json:"metadata"`
	Data     map[string]int    `json:"data"`
}

type TimeStruct struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type JSONTagStruct struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Ignored string `json:"-"`
	NoTag   string
}

type PointerStruct struct {
	Name    *string       `json:"name"`
	Address *SimpleStruct `json:"address"`
}

type ComplexStruct struct {
	Name      string            `json:"name"`
	Tags      []string          `json:"tags"`
	Addresses []NestedAddress   `json:"addresses"`
	Meta      map[string]string `json:"meta"`
	CreatedAt time.Time         `json:"created_at"`
}

type unexportedField struct {
	Public  string `json:"public"`
	private string //nolint:unused
}

type UintStruct struct {
	Count   uint   `json:"count"`
	Count8  uint8  `json:"count8"`
	Count16 uint16 `json:"count16"`
	Count32 uint32 `json:"count32"`
	Count64 uint64 `json:"count64"`
}

type IntVariantsStruct struct {
	I   int   `json:"i"`
	I8  int8  `json:"i8"`
	I16 int16 `json:"i16"`
	I32 int32 `json:"i32"`
	I64 int64 `json:"i64"`
}

type FloatVariantsStruct struct {
	F32 float32 `json:"f32"`
	F64 float64 `json:"f64"`
}

type DeeplyNested struct {
	Level1 struct {
		Level2 struct {
			Value string `json:"value"`
		} `json:"level2"`
	} `json:"level1"`
}

type ArrayOfStructs struct {
	Items []NestedAddress `json:"items"`
}

type MapOfObjects struct {
	Data map[string]NestedAddress `json:"data"`
}

// --- Structs for tag integration tests ---

type RequiredFields struct {
	Name  string `json:"name" required:"true"`
	Email string `json:"email" required:"true"`
	Age   int    `json:"age"`
}

type DefaultValues struct {
	Color    string  `json:"color" default:"blue"`
	Count    int     `json:"count" default:"10"`
	IsActive bool    `json:"is_active" default:"true"`
	Price    float64 `json:"price" default:"19.99"`
}

type EnumField struct {
	Status string `json:"status" enum:"active,inactive,pending"`
}

type FormatField struct {
	Email    string `json:"email" format:"email"`
	Birthday string `json:"birthday" format:"date"`
}

type FormatOverride struct {
	// format tag should override the auto-detected time.Time format.
	Created string `json:"created" format:"date"`
}

type AllTags struct {
	Email string `json:"email" required:"true" format:"email" default:"user@example.com" enum:"user@example.com,admin@example.com"`
}

type OmitemptyField struct {
	Name  string `json:"name"`
	Notes string `json:"notes,omitempty"`
}

// --- Structs for UI Schema tests ---

type UISimple struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
}

type UIWithLabel struct {
	Name string `json:"name" form:"label=Full name"`
}

type UIWithHidden struct {
	ID   int    `json:"id" form:"hidden"`
	Name string `json:"name"`
}

type UIWithReadonly struct {
	CreatedAt string `json:"created_at" form:"readonly"`
}

type UIWithMultiline struct {
	Bio string `json:"bio" form:"multiline"`
}

type UIWithCombinedForm struct {
	Name string `json:"name" form:"label=Full name;multiline;readonly"`
}

type UINestedAddr struct {
	City   string `json:"city"`
	Street string `json:"street"`
}

type UIWithNested struct {
	Name    string       `json:"name"`
	Address UINestedAddr `json:"address"`
}

type UINestedWithLabel struct {
	Name    string       `json:"name"`
	Address UINestedAddr `json:"address" form:"label=Home address"`
}

type UIDeeplyNested struct {
	Info struct {
		Contact struct {
			Phone string `json:"phone"`
		} `json:"contact"`
	} `json:"info"`
}

// --- Structs for rule tests ---

type UIWithVisibleIf struct {
	IsActive bool   `json:"is_active"`
	Details  string `json:"details" visibleIf:"is_active=true"`
}

type UIWithHideIf struct {
	Role   string `json:"role"`
	Secret string `json:"secret" hideIf:"role=admin"`
}

type UIWithEnableIf struct {
	Agreed bool   `json:"agreed"`
	Submit string `json:"submit" enableIf:"agreed=true"`
}

type UIWithDisableIf struct {
	Locked bool   `json:"locked"`
	Field  string `json:"field" disableIf:"locked=true"`
}

type UIWithIntRule struct {
	Level   int    `json:"level"`
	Details string `json:"details" visibleIf:"level=5"`
}

type UIWithStringRule struct {
	Status  string `json:"status"`
	Actions string `json:"actions" enableIf:"status=active"`
}

type UIWithMultipleRuleTags struct {
	Flag    bool   `json:"flag"`
	Content string `json:"content" visibleIf:"flag=true" hideIf:"flag=false"`
}

// --- Tests ---

func TestGenerateJSONSchema_SimpleStruct(t *testing.T) {
	s, err := parser.GenerateJSONSchema(SimpleStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Type, "object")

	expected := map[string]string{
		"name":      "string",
		"age":       "integer",
		"score":     "number",
		"is_active": "boolean",
	}

	for field, expectedType := range expected {
		prop, ok := s.Properties[field]
		if !ok {
			t.Errorf("missing property %q", field)
			continue
		}
		assertSchemaType(t, prop.Type, expectedType)
	}
}

func TestGenerateJSONSchema_Pointer(t *testing.T) {
	s, err := parser.GenerateJSONSchema(&SimpleStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSchemaType(t, s.Type, "object")

	if len(s.Properties) != 4 {
		t.Errorf("expected 4 properties, got %d", len(s.Properties))
	}
}

func TestGenerateJSONSchema_NestedStruct(t *testing.T) {
	s, err := parser.GenerateJSONSchema(NestedStruct{})
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

	cityProp, ok := addr.Properties["city"]
	if !ok {
		t.Fatal("missing nested property 'city'")
	}
	assertSchemaType(t, cityProp.Type, "string")
}

func TestGenerateJSONSchema_Slices(t *testing.T) {
	s, err := parser.GenerateJSONSchema(SliceStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tags, ok := s.Properties["tags"]
	if !ok {
		t.Fatal("missing property 'tags'")
	}
	assertSchemaType(t, tags.Type, "array")
	if tags.Items == nil || tags.Items.Type != typeString {
		t.Error("expected tags items to be 'string'")
	}

	scores, ok := s.Properties["scores"]
	if !ok {
		t.Fatal("missing property 'scores'")
	}
	assertSchemaType(t, scores.Type, "array")
	if scores.Items == nil || scores.Items.Type != "integer" {
		t.Error("expected scores items to be 'integer'")
	}
}

func TestGenerateJSONSchema_Maps(t *testing.T) {
	s, err := parser.GenerateJSONSchema(MapStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta, ok := s.Properties["metadata"]
	if !ok {
		t.Fatal("missing property 'metadata'")
	}
	assertSchemaType(t, meta.Type, "object")
	if meta.AdditionalProperties == nil || meta.AdditionalProperties.Type != typeString {
		t.Error("expected additionalProperties to be 'string'")
	}

	data, ok := s.Properties["data"]
	if !ok {
		t.Fatal("missing property 'data'")
	}
	assertSchemaType(t, data.Type, "object")
	if data.AdditionalProperties == nil || data.AdditionalProperties.Type != "integer" {
		t.Error("expected additionalProperties to be 'integer'")
	}
}

func TestGenerateJSONSchema_TimeFields(t *testing.T) {
	s, err := parser.GenerateJSONSchema(TimeStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	created, ok := s.Properties["created_at"]
	if !ok {
		t.Fatal("missing property 'created_at'")
	}
	assertSchemaType(t, created.Type, "string")
	if created.Format != "date-time" {
		t.Errorf("expected format 'date-time', got %q", created.Format)
	}

	updated, ok := s.Properties["updated_at"]
	if !ok {
		t.Fatal("missing property 'updated_at'")
	}
	assertSchemaType(t, updated.Type, "string")
	if updated.Format != "date-time" {
		t.Errorf("expected format 'date-time', got %q", updated.Format)
	}
}

func TestGenerateJSONSchema_JSONTags(t *testing.T) {
	s, err := parser.GenerateJSONSchema(JSONTagStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := s.Properties["id"]; !ok {
		t.Error("expected property 'id'")
	}

	if _, ok := s.Properties["name"]; !ok {
		t.Error("expected property 'name'")
	}

	if _, ok := s.Properties["Ignored"]; ok {
		t.Error("field with json:\"-\" should be excluded")
	}
	if _, ok := s.Properties["-"]; ok {
		t.Error("field with json:\"-\" should be excluded")
	}

	if _, ok := s.Properties["NoTag"]; !ok {
		t.Error("expected property 'NoTag' (no json tag)")
	}
}

func TestGenerateJSONSchema_PointerFields(t *testing.T) {
	s, err := parser.GenerateJSONSchema(PointerStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name, ok := s.Properties["name"]
	if !ok {
		t.Fatal("missing property 'name'")
	}
	assertSchemaType(t, name.Type, "string")

	addr, ok := s.Properties["address"]
	if !ok {
		t.Fatal("missing property 'address'")
	}
	assertSchemaType(t, addr.Type, "object")
}

func TestGenerateJSONSchema_UnexportedFieldsSkipped(t *testing.T) {
	s, err := parser.GenerateJSONSchema(unexportedField{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(s.Properties))
	}

	if _, ok := s.Properties["public"]; !ok {
		t.Error("expected property 'public'")
	}
}

func TestGenerateJSONSchema_ComplexStruct(t *testing.T) {
	s, err := parser.GenerateJSONSchema(ComplexStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Properties) != 5 {
		t.Errorf("expected 5 properties, got %d", len(s.Properties))
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("schema JSON is not valid: %v", err)
	}
}

func TestGenerateJSONSchema_UintTypes(t *testing.T) {
	s, err := parser.GenerateJSONSchema(UintStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, field := range []string{"count", "count8", "count16", "count32", "count64"} {
		prop, ok := s.Properties[field]
		if !ok {
			t.Errorf("missing property %q", field)
			continue
		}
		assertSchemaType(t, prop.Type, "integer")
	}
}

func TestGenerateJSONSchema_IntVariants(t *testing.T) {
	s, err := parser.GenerateJSONSchema(IntVariantsStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, field := range []string{"i", "i8", "i16", "i32", "i64"} {
		prop, ok := s.Properties[field]
		if !ok {
			t.Errorf("missing property %q", field)
			continue
		}
		assertSchemaType(t, prop.Type, "integer")
	}
}

func TestGenerateJSONSchema_FloatVariants(t *testing.T) {
	s, err := parser.GenerateJSONSchema(FloatVariantsStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, field := range []string{"f32", "f64"} {
		prop, ok := s.Properties[field]
		if !ok {
			t.Errorf("missing property %q", field)
			continue
		}
		assertSchemaType(t, prop.Type, "number")
	}
}

func TestGenerateJSONSchema_DeeplyNested(t *testing.T) {
	s, err := parser.GenerateJSONSchema(DeeplyNested{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l1, ok := s.Properties["level1"]
	if !ok {
		t.Fatal("missing property 'level1'")
	}
	assertSchemaType(t, l1.Type, "object")

	l2, ok := l1.Properties["level2"]
	if !ok {
		t.Fatal("missing property 'level2'")
	}
	assertSchemaType(t, l2.Type, "object")

	val, ok := l2.Properties["value"]
	if !ok {
		t.Fatal("missing property 'value'")
	}
	assertSchemaType(t, val.Type, "string")
}

func TestGenerateJSONSchema_ArrayOfStructs(t *testing.T) {
	s, err := parser.GenerateJSONSchema(ArrayOfStructs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, ok := s.Properties["items"]
	if !ok {
		t.Fatal("missing property 'items'")
	}
	assertSchemaType(t, items.Type, "array")

	if items.Items == nil {
		t.Fatal("expected items schema")
	}
	assertSchemaType(t, items.Items.Type, "object")

	if len(items.Items.Properties) != 2 {
		t.Errorf("expected 2 properties in array items, got %d", len(items.Items.Properties))
	}
}

func TestGenerateJSONSchema_MapOfObjects(t *testing.T) {
	s, err := parser.GenerateJSONSchema(MapOfObjects{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := s.Properties["data"]
	if !ok {
		t.Fatal("missing property 'data'")
	}
	assertSchemaType(t, data.Type, "object")

	if data.AdditionalProperties == nil {
		t.Fatal("expected additionalProperties schema")
	}
	assertSchemaType(t, data.AdditionalProperties.Type, "object")

	if len(data.AdditionalProperties.Properties) != 2 {
		t.Errorf("expected 2 properties in additionalProperties, got %d", len(data.AdditionalProperties.Properties))
	}
}

func TestGenerateJSONSchema_SchemaField(t *testing.T) {
	s, err := parser.GenerateJSONSchema(SimpleStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Schema != schemaDraft7 {
		t.Errorf("expected $schema to be set, got %q", s.Schema)
	}
}

func TestGenerateJSONSchema_ValidJSON(t *testing.T) {
	s, err := parser.GenerateJSONSchema(ComplexStruct{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !json.Valid(data) {
		t.Error("generated schema is not valid JSON")
	}
}

// --- Tag integration tests ---

func TestGenerateJSONSchema_RequiredTag(t *testing.T) {
	s, err := parser.GenerateJSONSchema(RequiredFields{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Required) != 2 {
		t.Fatalf("expected 2 required fields, got %d: %v", len(s.Required), s.Required)
	}

	requiredSet := make(map[string]bool)
	for _, r := range s.Required {
		requiredSet[r] = true
	}

	if !requiredSet["name"] {
		t.Error("expected 'name' in required")
	}
	if !requiredSet["email"] {
		t.Error("expected 'email' in required")
	}
	if requiredSet["age"] {
		t.Error("'age' should not be in required")
	}
}

func TestGenerateJSONSchema_DefaultTag(t *testing.T) {
	s, err := parser.GenerateJSONSchema(DefaultValues{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	color := s.Properties["color"]
	if color.Default != "blue" {
		t.Errorf("expected default 'blue', got %v", color.Default)
	}

	count := s.Properties["count"]
	if count.Default != int64(10) {
		t.Errorf("expected default 10, got %v (%T)", count.Default, count.Default)
	}

	active := s.Properties["is_active"]
	if active.Default != true {
		t.Errorf("expected default true, got %v", active.Default)
	}

	price := s.Properties["price"]
	if price.Default != 19.99 {
		t.Errorf("expected default 19.99, got %v", price.Default)
	}
}

func TestGenerateJSONSchema_EnumTag(t *testing.T) {
	s, err := parser.GenerateJSONSchema(EnumField{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status := s.Properties["status"]
	if len(status.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(status.Enum))
	}

	expected := []string{"active", "inactive", "pending"}
	for i, e := range expected {
		if status.Enum[i] != e {
			t.Errorf("enum[%d]: expected %q, got %v", i, e, status.Enum[i])
		}
	}
}

func TestGenerateJSONSchema_FormatTag(t *testing.T) {
	s, err := parser.GenerateJSONSchema(FormatField{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	email := s.Properties["email"]
	if email.Format != formatEmail {
		t.Errorf("expected format 'email', got %q", email.Format)
	}

	bday := s.Properties["birthday"]
	if bday.Format != "date" {
		t.Errorf("expected format 'date', got %q", bday.Format)
	}
}

func TestGenerateJSONSchema_FormatOverride(t *testing.T) {
	s, err := parser.GenerateJSONSchema(FormatOverride{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	created := s.Properties["created"]
	if created.Format != "date" {
		t.Errorf("expected format tag to set 'date', got %q", created.Format)
	}
}

func TestGenerateJSONSchema_AllTags(t *testing.T) {
	s, err := parser.GenerateJSONSchema(AllTags{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Required) != 1 || s.Required[0] != "email" {
		t.Errorf("expected required=['email'], got %v", s.Required)
	}

	email := s.Properties["email"]
	if email.Format != formatEmail {
		t.Errorf("expected format 'email', got %q", email.Format)
	}
	if email.Default != "user@example.com" {
		t.Errorf("expected default 'user@example.com', got %v", email.Default)
	}
	if len(email.Enum) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(email.Enum))
	}
}

func TestGenerateJSONSchema_OmitemptyField(t *testing.T) {
	s, err := parser.GenerateJSONSchema(OmitemptyField{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// omitempty fields should still appear in schema.
	if _, ok := s.Properties["name"]; !ok {
		t.Error("expected property 'name'")
	}
	if _, ok := s.Properties["notes"]; !ok {
		t.Error("expected property 'notes' (omitempty should still be in schema)")
	}
}

func TestGenerateJSONSchema_RequiredInJSON(t *testing.T) {
	s, err := parser.GenerateJSONSchema(RequiredFields{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	required, ok := result["required"].([]any)
	if !ok {
		t.Fatal("expected 'required' array in JSON output")
	}

	if len(required) != 2 {
		t.Errorf("expected 2 required fields in JSON, got %d", len(required))
	}
}

// --- UI Schema integration tests ---

func TestGenerateUISchema_Simple(t *testing.T) {
	ui, err := parser.GenerateUISchema(UISimple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(ui.Elements))
	}

	for _, el := range ui.Elements {
		if el.Type != "Control" {
			t.Errorf("expected Control, got %q", el.Type)
		}
		if el.Scope == "" {
			t.Error("expected non-empty scope")
		}
	}
}

func TestGenerateUISchema_Pointer(t *testing.T) {
	ui, err := parser.GenerateUISchema(&UISimple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ui.Type != typeVerticalLayout {
		t.Errorf("expected VerticalLayout, got %q", ui.Type)
	}

	if len(ui.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(ui.Elements))
	}
}

func TestGenerateUISchema_ScopeFormat(t *testing.T) {
	ui, err := parser.GenerateUISchema(UISimple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedScopes := map[string]bool{
		"#/properties/name":      false,
		"#/properties/age":       false,
		"#/properties/is_active": false,
	}

	for _, el := range ui.Elements {
		if _, ok := expectedScopes[el.Scope]; ok {
			expectedScopes[el.Scope] = true
		} else {
			t.Errorf("unexpected scope %q", el.Scope)
		}
	}

	for scope, found := range expectedScopes {
		if !found {
			t.Errorf("missing scope %q", scope)
		}
	}
}

func TestGenerateUISchema_Label(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithLabel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	el := ui.Elements[0]
	if el.Label != "Full name" {
		t.Errorf("expected label 'Full name', got %q", el.Label)
	}
}

func TestGenerateUISchema_Hidden(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithHidden{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hidden field should be excluded — only "name" remains.
	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element (hidden excluded), got %d", len(ui.Elements))
	}

	if ui.Elements[0].Scope != "#/properties/name" {
		t.Errorf("expected scope for 'name', got %q", ui.Elements[0].Scope)
	}
}

func TestGenerateUISchema_Readonly(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithReadonly{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	el := ui.Elements[0]
	if el.Options == nil {
		t.Fatal("expected options to be set")
	}
	if el.Options["readonly"] != true {
		t.Errorf("expected readonly=true, got %v", el.Options["readonly"])
	}
}

func TestGenerateUISchema_Multiline(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithMultiline{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	el := ui.Elements[0]
	if el.Options == nil {
		t.Fatal("expected options to be set")
	}
	if el.Options["multi"] != true {
		t.Errorf("expected multi=true, got %v", el.Options["multi"])
	}
}

func TestGenerateUISchema_CombinedFormOptions(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithCombinedForm{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(ui.Elements))
	}

	el := ui.Elements[0]
	if el.Label != "Full name" {
		t.Errorf("expected label 'Full name', got %q", el.Label)
	}
	if el.Options == nil {
		t.Fatal("expected options to be set")
	}
	if el.Options["multi"] != true {
		t.Errorf("expected multi=true")
	}
	if el.Options["readonly"] != true {
		t.Errorf("expected readonly=true")
	}
}

func TestGenerateUISchema_NestedStruct(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithNested{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: Control(name) + Group(address)
	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	// Find the group element.
	var group *parserTestElement
	for _, el := range ui.Elements {
		if el.Type == typeGroup {
			group = &parserTestElement{el.Type, el.Label, el.Scope, len(el.Elements)}
			break
		}
	}

	if group == nil {
		t.Fatal("expected a Group element for nested struct")
	}

	if group.label != "Address" {
		t.Errorf("expected group label 'Address', got %q", group.label)
	}

	if group.childCount != 2 {
		t.Errorf("expected 2 children in group, got %d", group.childCount)
	}
}

func TestGenerateUISchema_NestedWithLabel(t *testing.T) {
	ui, err := parser.GenerateUISchema(UINestedWithLabel{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var group *parserTestElement
	for _, el := range ui.Elements {
		if el.Type == typeGroup {
			group = &parserTestElement{el.Type, el.Label, el.Scope, len(el.Elements)}
			break
		}
	}

	if group == nil {
		t.Fatal("expected a Group element")
	}

	if group.label != "Home address" {
		t.Errorf("expected group label 'Home address', got %q", group.label)
	}
}

func TestGenerateUISchema_DeeplyNested(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIDeeplyNested{})
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
	if phone.Type != "Control" {
		t.Errorf("expected Control, got %q", phone.Type)
	}
	if phone.Scope != "#/properties/info/properties/contact/properties/phone" {
		t.Errorf("unexpected scope %q", phone.Scope)
	}
}

func TestGenerateUISchema_ValidJSON(t *testing.T) {
	ui, err := parser.GenerateUISchema(UISimple{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(ui)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !json.Valid(data) {
		t.Error("generated UI schema is not valid JSON")
	}
}

// --- Rule integration tests ---

func TestGenerateUISchema_VisibleIf(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithVisibleIf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ui.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(ui.Elements))
	}

	// First element (is_active) should have no rule.
	if ui.Elements[0].Rule != nil {
		t.Error("expected no rule on is_active control")
	}

	// Second element (details) should have SHOW rule.
	details := ui.Elements[1]
	if details.Rule == nil {
		t.Fatal("expected rule on details control")
	}
	if details.Rule.Effect != effectShow {
		t.Errorf("expected effect SHOW, got %q", details.Rule.Effect)
	}
	if details.Rule.Condition.Scope != "#/properties/is_active" {
		t.Errorf("expected condition scope '#/properties/is_active', got %q", details.Rule.Condition.Scope)
	}
	if details.Rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", details.Rule.Condition.Schema.Const)
	}
}

func TestGenerateUISchema_HideIf(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithHideIf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := ui.Elements[1]
	if secret.Rule == nil {
		t.Fatal("expected rule on secret control")
	}
	if secret.Rule.Effect != "HIDE" {
		t.Errorf("expected effect HIDE, got %q", secret.Rule.Effect)
	}
	if secret.Rule.Condition.Schema.Const != "admin" {
		t.Errorf("expected const 'admin', got %v", secret.Rule.Condition.Schema.Const)
	}
}

func TestGenerateUISchema_EnableIf(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithEnableIf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	submit := ui.Elements[1]
	if submit.Rule == nil {
		t.Fatal("expected rule on submit control")
	}
	if submit.Rule.Effect != "ENABLE" {
		t.Errorf("expected effect ENABLE, got %q", submit.Rule.Effect)
	}
}

func TestGenerateUISchema_DisableIf(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithDisableIf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	field := ui.Elements[1]
	if field.Rule == nil {
		t.Fatal("expected rule on field control")
	}
	if field.Rule.Effect != "DISABLE" {
		t.Errorf("expected effect DISABLE, got %q", field.Rule.Effect)
	}
}

func TestGenerateUISchema_RuleWithIntValue(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithIntRule{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	details := ui.Elements[1]
	if details.Rule == nil {
		t.Fatal("expected rule on details control")
	}
	if details.Rule.Condition.Schema.Const != int64(5) {
		t.Errorf("expected const 5 (int64), got %v (%T)", details.Rule.Condition.Schema.Const, details.Rule.Condition.Schema.Const)
	}
}

func TestGenerateUISchema_RuleWithStringValue(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithStringRule{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions := ui.Elements[1]
	if actions.Rule == nil {
		t.Fatal("expected rule on actions control")
	}
	if actions.Rule.Effect != "ENABLE" {
		t.Errorf("expected effect ENABLE, got %q", actions.Rule.Effect)
	}
	if actions.Rule.Condition.Schema.Const != "active" {
		t.Errorf("expected const 'active', got %v", actions.Rule.Condition.Schema.Const)
	}
}

func TestGenerateUISchema_MultipleRuleTags_FirstWins(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithMultipleRuleTags{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := ui.Elements[1]
	if content.Rule == nil {
		t.Fatal("expected rule on content control")
	}
	// visibleIf has priority over hideIf.
	if content.Rule.Effect != effectShow {
		t.Errorf("expected effect SHOW (visibleIf takes priority), got %q", content.Rule.Effect)
	}
}

func TestGenerateUISchema_RuleValidJSON(t *testing.T) {
	ui, err := parser.GenerateUISchema(UIWithVisibleIf{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(ui)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !json.Valid(data) {
		t.Error("generated UI schema with rules is not valid JSON")
	}

	// Verify JSON structure contains rule.
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	elements, ok := result["elements"].([]any)
	if !ok {
		t.Fatal("expected elements array")
	}

	detailsEl, ok := elements[1].(map[string]any)
	if !ok {
		t.Fatal("expected details element as object")
	}

	rule, ok := detailsEl["rule"].(map[string]any)
	if !ok {
		t.Fatal("expected rule in details element")
	}

	if rule["effect"] != effectShow {
		t.Errorf("expected effect SHOW in JSON, got %v", rule["effect"])
	}
}

// --- helpers ---

type parserTestElement struct {
	typ        string
	label      string
	scope      string
	childCount int
}

func assertSchemaType(t *testing.T, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("expected type %q, got %q", expected, got)
	}
}
