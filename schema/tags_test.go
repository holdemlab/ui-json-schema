package schema_test

import (
	"reflect"
	"testing"

	"github.com/holdemlab/ui-json-schema/schema"
)

type tagRequired struct {
	Name string `required:"true"`
}

type tagRequiredFalse struct {
	Name string `required:"false"`
}

type tagDefaultString struct {
	Color string `default:"red"`
}

type tagDefaultBool struct {
	Active bool `default:"true"`
}

type tagDefaultInt struct {
	Count int `default:"42"`
}

type tagDefaultFloat struct {
	Price float64 `default:"9.99"`
}

type tagDefaultUint struct {
	Size uint `default:"10"`
}

type tagDefaultInvalid struct {
	Count int `default:"not_a_number"`
}

type tagEnum struct {
	Status string `enum:"active,inactive,pending"`
}

type tagEnumSpaces struct {
	Status string `enum:"a , b , c"`
}

type tagFormat struct {
	Email string `format:"email"`
}

type tagFormatDate struct {
	Birthday string `format:"date"`
}

type tagForm struct {
	Name string `form:"label=Full name;multiline"`
}

type tagVisibleIf struct {
	Details string `visibleIf:"is_active=true"`
}

type tagHideIf struct {
	Secret string `hideIf:"role=admin"`
}

type tagEnableIf struct {
	Submit string `enableIf:"agreed=true"`
}

type tagDisableIf struct {
	Field string `disableIf:"locked=true"`
}

type tagCombined struct {
	Email string `json:"email" required:"true" format:"email" default:"user@example.com"`
}

func TestParseFieldTags_Required(t *testing.T) {
	field, _ := reflect.TypeOf(tagRequired{}).FieldByName("Name")
	tags := schema.ParseFieldTags(field)
	if !tags.Required {
		t.Error("expected Required to be true")
	}
}

func TestParseFieldTags_RequiredFalse(t *testing.T) {
	field, _ := reflect.TypeOf(tagRequiredFalse{}).FieldByName("Name")
	tags := schema.ParseFieldTags(field)
	if tags.Required {
		t.Error("expected Required to be false")
	}
}

func TestParseFieldTags_DefaultString(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultString{}).FieldByName("Color")
	tags := schema.ParseFieldTags(field)
	if tags.Default != "red" {
		t.Errorf("expected default 'red', got %v", tags.Default)
	}
}

func TestParseFieldTags_DefaultBool(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultBool{}).FieldByName("Active")
	tags := schema.ParseFieldTags(field)
	if tags.Default != true {
		t.Errorf("expected default true, got %v", tags.Default)
	}
}

func TestParseFieldTags_DefaultInt(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultInt{}).FieldByName("Count")
	tags := schema.ParseFieldTags(field)
	if tags.Default != int64(42) {
		t.Errorf("expected default 42, got %v (%T)", tags.Default, tags.Default)
	}
}

func TestParseFieldTags_DefaultFloat(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultFloat{}).FieldByName("Price")
	tags := schema.ParseFieldTags(field)
	if tags.Default != 9.99 {
		t.Errorf("expected default 9.99, got %v", tags.Default)
	}
}

func TestParseFieldTags_DefaultUint(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultUint{}).FieldByName("Size")
	tags := schema.ParseFieldTags(field)
	if tags.Default != uint64(10) {
		t.Errorf("expected default 10, got %v (%T)", tags.Default, tags.Default)
	}
}

func TestParseFieldTags_DefaultInvalidInt(t *testing.T) {
	field, _ := reflect.TypeOf(tagDefaultInvalid{}).FieldByName("Count")
	tags := schema.ParseFieldTags(field)
	if tags.Default != "not_a_number" {
		t.Errorf("expected fallback string default, got %v", tags.Default)
	}
}

func TestParseFieldTags_Enum(t *testing.T) {
	field, _ := reflect.TypeOf(tagEnum{}).FieldByName("Status")
	tags := schema.ParseFieldTags(field)
	if len(tags.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(tags.Enum))
	}
	expected := []string{"active", "inactive", "pending"}
	for i, e := range expected {
		if tags.Enum[i] != e {
			t.Errorf("enum[%d]: expected %q, got %v", i, e, tags.Enum[i])
		}
	}
}

func TestParseFieldTags_EnumSpaces(t *testing.T) {
	field, _ := reflect.TypeOf(tagEnumSpaces{}).FieldByName("Status")
	tags := schema.ParseFieldTags(field)
	if len(tags.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(tags.Enum))
	}
	expected := []string{"a", "b", "c"}
	for i, e := range expected {
		if tags.Enum[i] != e {
			t.Errorf("enum[%d]: expected %q, got %v", i, e, tags.Enum[i])
		}
	}
}

func TestParseFieldTags_Format(t *testing.T) {
	field, _ := reflect.TypeOf(tagFormat{}).FieldByName("Email")
	tags := schema.ParseFieldTags(field)
	if tags.Format != "email" {
		t.Errorf("expected format 'email', got %q", tags.Format)
	}
}

func TestParseFieldTags_FormatDate(t *testing.T) {
	field, _ := reflect.TypeOf(tagFormatDate{}).FieldByName("Birthday")
	tags := schema.ParseFieldTags(field)
	if tags.Format != "date" {
		t.Errorf("expected format 'date', got %q", tags.Format)
	}
}

func TestParseFieldTags_Form(t *testing.T) {
	field, _ := reflect.TypeOf(tagForm{}).FieldByName("Name")
	tags := schema.ParseFieldTags(field)
	if tags.Form != "label=Full name;multiline" {
		t.Errorf("expected form tag value, got %q", tags.Form)
	}
}

func TestParseFieldTags_VisibleIf(t *testing.T) {
	field, _ := reflect.TypeOf(tagVisibleIf{}).FieldByName("Details")
	tags := schema.ParseFieldTags(field)
	if tags.VisibleIf != "is_active=true" {
		t.Errorf("expected visibleIf tag value, got %q", tags.VisibleIf)
	}
}

func TestParseFieldTags_HideIf(t *testing.T) {
	field, _ := reflect.TypeOf(tagHideIf{}).FieldByName("Secret")
	tags := schema.ParseFieldTags(field)
	if tags.HideIf != "role=admin" {
		t.Errorf("expected hideIf 'role=admin', got %q", tags.HideIf)
	}
}

func TestParseFieldTags_EnableIf(t *testing.T) {
	field, _ := reflect.TypeOf(tagEnableIf{}).FieldByName("Submit")
	tags := schema.ParseFieldTags(field)
	if tags.EnableIf != "agreed=true" {
		t.Errorf("expected enableIf 'agreed=true', got %q", tags.EnableIf)
	}
}

func TestParseFieldTags_DisableIf(t *testing.T) {
	field, _ := reflect.TypeOf(tagDisableIf{}).FieldByName("Field")
	tags := schema.ParseFieldTags(field)
	if tags.DisableIf != "locked=true" {
		t.Errorf("expected disableIf 'locked=true', got %q", tags.DisableIf)
	}
}

func TestParseFieldTags_Combined(t *testing.T) {
	field, _ := reflect.TypeOf(tagCombined{}).FieldByName("Email")
	tags := schema.ParseFieldTags(field)
	if !tags.Required {
		t.Error("expected Required to be true")
	}
	if tags.Format != "email" {
		t.Errorf("expected format 'email', got %q", tags.Format)
	}
	if tags.Default != "user@example.com" {
		t.Errorf("expected default 'user@example.com', got %v", tags.Default)
	}
}

func TestParseFieldTags_NoTags(t *testing.T) {
	type noTags struct {
		Name string
	}
	field, _ := reflect.TypeOf(noTags{}).FieldByName("Name")
	tags := schema.ParseFieldTags(field)
	if tags.Required {
		t.Error("expected Required to be false")
	}
	if tags.Default != nil {
		t.Errorf("expected nil Default, got %v", tags.Default)
	}
	if len(tags.Enum) != 0 {
		t.Errorf("expected empty Enum, got %v", tags.Enum)
	}
	if tags.Format != "" {
		t.Errorf("expected empty Format, got %q", tags.Format)
	}
}

func TestParseFieldTags_I18nKey(t *testing.T) {
	type i18nStruct struct {
		Name string `i18n:"user.name"`
	}

	field, _ := reflect.TypeOf(i18nStruct{}).FieldByName("Name")
	tags := schema.ParseFieldTags(field)

	if tags.I18nKey != "user.name" {
		t.Errorf("expected I18nKey 'user.name', got %q", tags.I18nKey)
	}
}

func TestParseFieldTags_Renderer(t *testing.T) {
	type rendererStruct struct {
		Color string `renderer:"color-picker"`
	}

	field, _ := reflect.TypeOf(rendererStruct{}).FieldByName("Color")
	tags := schema.ParseFieldTags(field)

	if tags.Renderer != "color-picker" {
		t.Errorf("expected Renderer 'color-picker', got %q", tags.Renderer)
	}
}
