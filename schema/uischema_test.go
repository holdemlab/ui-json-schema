package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/holdemlab/ui-json-schema/schema"
)

const (
	labelFullName = "Full name"
	scopeIsActive = "#/properties/is_active"
)

func TestNewVerticalLayout(t *testing.T) {
	vl := schema.NewVerticalLayout()
	if vl.Type != "VerticalLayout" {
		t.Errorf("expected type 'VerticalLayout', got %q", vl.Type)
	}
	if vl.Elements == nil {
		t.Error("expected non-nil elements slice")
	}
	if len(vl.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(vl.Elements))
	}
}

func TestNewHorizontalLayout(t *testing.T) {
	hl := schema.NewHorizontalLayout()
	if hl.Type != "HorizontalLayout" {
		t.Errorf("expected type 'HorizontalLayout', got %q", hl.Type)
	}
}

func TestNewGroup(t *testing.T) {
	g := schema.NewGroup("Address")
	if g.Type != "Group" {
		t.Errorf("expected type 'Group', got %q", g.Type)
	}
	if g.Label != "Address" {
		t.Errorf("expected label 'Address', got %q", g.Label)
	}
	if g.Elements == nil {
		t.Error("expected non-nil elements slice")
	}
}

func TestNewControl(t *testing.T) {
	c := schema.NewControl("#/properties/name")
	if c.Type != "Control" {
		t.Errorf("expected type 'Control', got %q", c.Type)
	}
	if c.Scope != "#/properties/name" {
		t.Errorf("expected scope '#/properties/name', got %q", c.Scope)
	}
}

func TestUISchemaElement_MarshalJSON(t *testing.T) {
	vl := schema.NewVerticalLayout()
	vl.Elements = append(vl.Elements,
		schema.NewControl("#/properties/name"),
		schema.NewControl("#/properties/age"),
	)

	data, err := json.Marshal(vl)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result["type"] != "VerticalLayout" {
		t.Errorf("expected type 'VerticalLayout', got %v", result["type"])
	}

	elements, ok := result["elements"].([]any)
	if !ok {
		t.Fatal("expected elements array")
	}
	if len(elements) != 2 {
		t.Errorf("expected 2 elements, got %d", len(elements))
	}
}

func TestUISchemaElement_OmitEmpty(t *testing.T) {
	c := schema.NewControl("#/properties/name")

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, field := range []string{"label", "elements", "options", "rule"} {
		if _, ok := result[field]; ok {
			t.Errorf("expected field %q to be omitted", field)
		}
	}
}

func TestUISchemaElement_WithOptions(t *testing.T) {
	c := schema.NewControl("#/properties/bio")
	c.Options = map[string]any{
		"multi": true,
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	opts, ok := result["options"].(map[string]any)
	if !ok {
		t.Fatal("expected options object")
	}
	if opts["multi"] != true {
		t.Errorf("expected options.multi=true, got %v", opts["multi"])
	}
}

func TestParseFormTag_Empty(t *testing.T) {
	opts := schema.ParseFormTag("")
	if opts.Label != "" || opts.Hidden || opts.Readonly || opts.Multiline {
		t.Error("expected all defaults for empty tag")
	}
}

func TestParseFormTag_Label(t *testing.T) {
	opts := schema.ParseFormTag("label=Full name")
	if opts.Label != labelFullName {
		t.Errorf("expected label 'Full name', got %q", opts.Label)
	}
}

func TestParseFormTag_Hidden(t *testing.T) {
	opts := schema.ParseFormTag("hidden")
	if !opts.Hidden {
		t.Error("expected Hidden to be true")
	}
}

func TestParseFormTag_Readonly(t *testing.T) {
	opts := schema.ParseFormTag("readonly")
	if !opts.Readonly {
		t.Error("expected Readonly to be true")
	}
}

func TestParseFormTag_Multiline(t *testing.T) {
	opts := schema.ParseFormTag("multiline")
	if !opts.Multiline {
		t.Error("expected Multiline to be true")
	}
}

func TestParseFormTag_Combined(t *testing.T) {
	opts := schema.ParseFormTag("label=Full name;multiline;readonly")
	if opts.Label != labelFullName {
		t.Errorf("expected label 'Full name', got %q", opts.Label)
	}
	if !opts.Multiline {
		t.Error("expected Multiline to be true")
	}
	if !opts.Readonly {
		t.Error("expected Readonly to be true")
	}
	if opts.Hidden {
		t.Error("expected Hidden to be false")
	}
}

func TestParseFormTag_Spaces(t *testing.T) {
	opts := schema.ParseFormTag(" label = Full name ; multiline ")
	if opts.Label != labelFullName {
		t.Errorf("expected label 'Full name', got %q", opts.Label)
	}
	if !opts.Multiline {
		t.Error("expected Multiline to be true")
	}
}

func TestUISchemaRule_MarshalJSON(t *testing.T) {
	c := schema.NewControl("#/properties/details")
	c.Rule = &schema.UISchemaRule{
		Effect: "SHOW",
		Condition: &schema.UISchemaCondition{
			Scope:  scopeIsActive,
			Schema: &schema.JSONSchema{Const: true},
		},
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	rule, ok := result["rule"].(map[string]any)
	if !ok {
		t.Fatal("expected rule object")
	}
	if rule["effect"] != "SHOW" {
		t.Errorf("expected effect 'SHOW', got %v", rule["effect"])
	}

	cond, ok := rule["condition"].(map[string]any)
	if !ok {
		t.Fatal("expected condition object")
	}
	if cond["scope"] != scopeIsActive {
		t.Errorf("expected condition scope, got %v", cond["scope"])
	}
}

func TestParseRuleExpression_BoolValue(t *testing.T) {
	rule := schema.ParseRuleExpression("is_active=true", schema.EffectShow)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Effect != schema.EffectShow {
		t.Errorf("expected effect SHOW, got %q", rule.Effect)
	}
	if rule.Condition.Scope != scopeIsActive {
		t.Errorf("expected scope '#/properties/is_active', got %q", rule.Condition.Scope)
	}
	if rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", rule.Condition.Schema.Const)
	}
}

func TestParseRuleExpression_StringValue(t *testing.T) {
	rule := schema.ParseRuleExpression("role=admin", schema.EffectEnable)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Effect != schema.EffectEnable {
		t.Errorf("expected effect ENABLE, got %q", rule.Effect)
	}
	if rule.Condition.Schema.Const != "admin" {
		t.Errorf("expected const 'admin', got %v", rule.Condition.Schema.Const)
	}
}

func TestParseRuleExpression_IntValue(t *testing.T) {
	rule := schema.ParseRuleExpression("count=42", schema.EffectHide)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Condition.Schema.Const != int64(42) {
		t.Errorf("expected const 42, got %v (%T)", rule.Condition.Schema.Const, rule.Condition.Schema.Const)
	}
}

func TestParseRuleExpression_FloatValue(t *testing.T) {
	rule := schema.ParseRuleExpression("price=9.99", schema.EffectDisable)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Condition.Schema.Const != 9.99 {
		t.Errorf("expected const 9.99, got %v", rule.Condition.Schema.Const)
	}
}

func TestParseRuleExpression_Empty(t *testing.T) {
	rule := schema.ParseRuleExpression("", schema.EffectShow)
	if rule != nil {
		t.Error("expected nil rule for empty expression")
	}
}

func TestParseRuleExpression_NoEquals(t *testing.T) {
	rule := schema.ParseRuleExpression("invalid", schema.EffectShow)
	if rule != nil {
		t.Error("expected nil rule for expression without '='")
	}
}

func TestParseRuleExpression_EmptyField(t *testing.T) {
	rule := schema.ParseRuleExpression("=value", schema.EffectShow)
	if rule != nil {
		t.Error("expected nil rule for empty field name")
	}
}

func TestParseRuleExpression_Spaces(t *testing.T) {
	rule := schema.ParseRuleExpression(" is_active = true ", schema.EffectShow)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Condition.Scope != scopeIsActive {
		t.Errorf("expected trimmed scope, got %q", rule.Condition.Scope)
	}
	if rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", rule.Condition.Schema.Const)
	}
}

func TestParseRuleExpression_FalseValue(t *testing.T) {
	rule := schema.ParseRuleExpression("enabled=false", schema.EffectHide)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}
	if rule.Effect != schema.EffectHide {
		t.Errorf("expected effect HIDE, got %q", rule.Effect)
	}
	if rule.Condition.Schema.Const != false {
		t.Errorf("expected const false, got %v", rule.Condition.Schema.Const)
	}
}

func TestNewCategorization(t *testing.T) {
	cat := schema.NewCategorization()
	if cat.Type != "Categorization" {
		t.Errorf("expected type 'Categorization', got %q", cat.Type)
	}

	if cat.Elements == nil {
		t.Error("expected non-nil elements slice")
	}

	if len(cat.Elements) != 0 {
		t.Errorf("expected 0 elements, got %d", len(cat.Elements))
	}
}

func TestNewCategory(t *testing.T) {
	cat := schema.NewCategory("Personal")
	if cat.Type != "Category" {
		t.Errorf("expected type 'Category', got %q", cat.Type)
	}

	if cat.Label != "Personal" {
		t.Errorf("expected label 'Personal', got %q", cat.Label)
	}

	if cat.Elements == nil {
		t.Error("expected non-nil elements slice")
	}
}

func TestParseFormTag_Category(t *testing.T) {
	opts := schema.ParseFormTag("category=Personal Info")
	if opts.Category != "Personal Info" {
		t.Errorf("expected category 'Personal Info', got %q", opts.Category)
	}
}

func TestParseFormTag_Layout(t *testing.T) {
	opts := schema.ParseFormTag("layout=horizontal")
	if opts.Layout != "horizontal" {
		t.Errorf("expected layout 'horizontal', got %q", opts.Layout)
	}
}

func TestParseFormTag_CategoryWithOtherOptions(t *testing.T) {
	opts := schema.ParseFormTag("label=Name;category=Basic;readonly")
	if opts.Label != "Name" {
		t.Errorf("expected label 'Name', got %q", opts.Label)
	}

	if opts.Category != "Basic" {
		t.Errorf("expected category 'Basic', got %q", opts.Category)
	}

	if !opts.Readonly {
		t.Error("expected readonly to be true")
	}
}

func TestParseFormTag_VisibleIf(t *testing.T) {
	opts := schema.ParseFormTag("category=Address;visibleIf=provideAddress:true")
	if opts.Category != "Address" {
		t.Errorf("expected category 'Address', got %q", opts.Category)
	}

	if opts.VisibleIf != "provideAddress:true" {
		t.Errorf("expected visibleIf 'provideAddress:true', got %q", opts.VisibleIf)
	}
}

func TestParseFormTag_HideIf(t *testing.T) {
	opts := schema.ParseFormTag("category=Secret;hideIf=role:admin")
	if opts.HideIf != "role:admin" {
		t.Errorf("expected hideIf 'role:admin', got %q", opts.HideIf)
	}
}

func TestParseFormTag_EnableIf(t *testing.T) {
	opts := schema.ParseFormTag("enableIf=active:true")
	if opts.EnableIf != "active:true" {
		t.Errorf("expected enableIf 'active:true', got %q", opts.EnableIf)
	}
}

func TestParseFormTag_DisableIf(t *testing.T) {
	opts := schema.ParseFormTag("disableIf=locked:true")
	if opts.DisableIf != "locked:true" {
		t.Errorf("expected disableIf 'locked:true', got %q", opts.DisableIf)
	}
}

func TestParseFormRuleExpression_Show(t *testing.T) {
	rule := schema.ParseFormRuleExpression("provideAddress:true", schema.EffectShow)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.Effect != schema.EffectShow {
		t.Errorf("expected effect SHOW, got %q", rule.Effect)
	}

	if rule.Condition.Scope != "#/properties/provideAddress" {
		t.Errorf("expected scope '#/properties/provideAddress', got %q", rule.Condition.Scope)
	}

	if rule.Condition.Schema.Const != true {
		t.Errorf("expected const true, got %v", rule.Condition.Schema.Const)
	}
}

func TestParseFormRuleExpression_Hide(t *testing.T) {
	rule := schema.ParseFormRuleExpression("status:draft", schema.EffectHide)
	if rule == nil {
		t.Fatal("expected non-nil rule")
	}

	if rule.Effect != schema.EffectHide {
		t.Errorf("expected effect HIDE, got %q", rule.Effect)
	}

	if rule.Condition.Schema.Const != "draft" {
		t.Errorf("expected const 'draft', got %v", rule.Condition.Schema.Const)
	}
}

func TestParseFormRuleExpression_Empty(t *testing.T) {
	rule := schema.ParseFormRuleExpression("", schema.EffectShow)
	if rule != nil {
		t.Error("expected nil rule for empty expression")
	}
}

func TestParseFormTag_I18n(t *testing.T) {
	opts := schema.ParseFormTag("category=Personal;i18n=category.personal")
	if opts.Category != "Personal" {
		t.Errorf("expected category 'Personal', got %q", opts.Category)
	}

	if opts.I18nKey != "category.personal" {
		t.Errorf("expected i18n key 'category.personal', got %q", opts.I18nKey)
	}
}

func TestParseFormTag_I18nAlone(t *testing.T) {
	opts := schema.ParseFormTag("i18n=some.key")
	if opts.I18nKey != "some.key" {
		t.Errorf("expected i18n key 'some.key', got %q", opts.I18nKey)
	}
}

func TestUISchemaElement_I18nJSON(t *testing.T) {
	el := &schema.UISchemaElement{
		Type:  "Category",
		Label: "Personal",
		I18n:  "category.personal",
	}

	data, err := json.Marshal(el)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	if parsed["i18n"] != "category.personal" {
		t.Errorf("expected i18n 'category.personal' in JSON, got %v", parsed["i18n"])
	}
}

func TestUISchemaElement_I18nOmitEmpty(t *testing.T) {
	el := &schema.UISchemaElement{
		Type:  "Category",
		Label: "Personal",
	}

	data, err := json.Marshal(el)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	if _, ok := parsed["i18n"]; ok {
		t.Error("expected i18n to be omitted when empty")
	}
}
