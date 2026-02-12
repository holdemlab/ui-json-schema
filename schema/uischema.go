package schema

import (
	"strconv"
	"strings"
)

// UISchemaElement represents a single element in a JSON Forms UI Schema.
// It can be a layout (VerticalLayout, HorizontalLayout, Group) or a Control.
type UISchemaElement struct {
	Type     string             `json:"type"`
	Label    string             `json:"label,omitempty"`
	I18n     string             `json:"i18n,omitempty"`
	Scope    string             `json:"scope,omitempty"`
	Elements []*UISchemaElement `json:"elements,omitempty"`
	Options  map[string]any     `json:"options,omitempty"`
	Rule     *UISchemaRule      `json:"rule,omitempty"`
}

// UISchemaRule represents a conditional visibility/enable rule in JSON Forms.
type UISchemaRule struct {
	Effect    string             `json:"effect"`
	Condition *UISchemaCondition `json:"condition"`
}

// UISchemaCondition represents the condition part of a UI Schema rule.
type UISchemaCondition struct {
	Scope  string      `json:"scope"`
	Schema *JSONSchema `json:"schema"`
}

// Rule effects.
const (
	EffectShow    = "SHOW"
	EffectHide    = "HIDE"
	EffectEnable  = "ENABLE"
	EffectDisable = "DISABLE"
)

// NewVerticalLayout creates a new VerticalLayout UI Schema element.
func NewVerticalLayout() *UISchemaElement {
	return &UISchemaElement{
		Type:     "VerticalLayout",
		Elements: make([]*UISchemaElement, 0),
	}
}

// NewHorizontalLayout creates a new HorizontalLayout UI Schema element.
func NewHorizontalLayout() *UISchemaElement {
	return &UISchemaElement{
		Type:     "HorizontalLayout",
		Elements: make([]*UISchemaElement, 0),
	}
}

// NewGroup creates a new Group UI Schema element with a label.
func NewGroup(label string) *UISchemaElement {
	return &UISchemaElement{
		Type:     "Group",
		Label:    label,
		Elements: make([]*UISchemaElement, 0),
	}
}

// NewCategorization creates a new Categorization UI Schema element.
// Categorization is a top-level layout that contains Category children,
// each rendered as a tab or wizard step by JSON Forms.
func NewCategorization() *UISchemaElement {
	return &UISchemaElement{
		Type:     "Categorization",
		Elements: make([]*UISchemaElement, 0),
	}
}

// NewCategory creates a new Category element with a label.
// Categories are children of a Categorization layout.
func NewCategory(label string) *UISchemaElement {
	return &UISchemaElement{
		Type:     "Category",
		Label:    label,
		Elements: make([]*UISchemaElement, 0),
	}
}

// NewControl creates a new Control element pointing to the given JSON path.
func NewControl(scope string) *UISchemaElement {
	return &UISchemaElement{
		Type:  "Control",
		Scope: scope,
	}
}

// FormOptions holds parsed form tag metadata for UI Schema generation.
type FormOptions struct {
	Label     string
	Hidden    bool
	Readonly  bool
	Multiline bool
	// Category assigns the field to a named category (tab) inside a Categorization layout.
	Category string
	// Layout overrides the default VerticalLayout. Supported: "horizontal".
	Layout string
	// VisibleIf holds a SHOW rule expression for the parent layout (Category/Group).
	// Format: "field:value" (colon-separated because '=' is used by the form tag syntax).
	VisibleIf string
	// HideIf holds a HIDE rule expression for the parent layout.
	HideIf string
	// EnableIf holds an ENABLE rule expression for the parent layout.
	EnableIf string
	// DisableIf holds a DISABLE rule expression for the parent layout.
	DisableIf string
	// I18nKey holds an i18n translation key for the Category label.
	I18nKey string
}

// ParseFormTag parses a form struct tag value like "label=Full name;multiline;readonly".
func ParseFormTag(tag string) FormOptions {
	var opts FormOptions

	if tag == "" {
		return opts
	}

	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value, hasValue := strings.Cut(part, "=")
		key = strings.TrimSpace(key)

		switch key {
		case "label":
			if hasValue {
				opts.Label = strings.TrimSpace(value)
			}
		case "hidden":
			opts.Hidden = true
		case "readonly":
			opts.Readonly = true
		case "multiline":
			opts.Multiline = true
		case "category":
			if hasValue {
				opts.Category = strings.TrimSpace(value)
			}
		case "layout":
			if hasValue {
				opts.Layout = strings.TrimSpace(value)
			}
		case "i18n":
			if hasValue {
				opts.I18nKey = strings.TrimSpace(value)
			}
		default:
			parseFormRulePart(key, value, hasValue, &opts)
		}
	}

	return opts
}

// parseFormRulePart handles rule-related keys inside a form tag
// (visibleIf, hideIf, enableIf, disableIf).
func parseFormRulePart(key, value string, hasValue bool, opts *FormOptions) {
	if !hasValue {
		return
	}

	v := strings.TrimSpace(value)

	switch key {
	case "visibleIf":
		opts.VisibleIf = v
	case "hideIf":
		opts.HideIf = v
	case "enableIf":
		opts.EnableIf = v
	case "disableIf":
		opts.DisableIf = v
	}
}

// ParseRuleExpression parses a condition expression like "field=value" and returns
// a UISchemaRule with the given effect. The scope is built relative to
// #/properties/<field>.
func ParseRuleExpression(expr string, effect string) *UISchemaRule {
	if expr == "" {
		return nil
	}

	field, rawValue, ok := strings.Cut(expr, "=")
	if !ok {
		return nil
	}

	field = strings.TrimSpace(field)
	rawValue = strings.TrimSpace(rawValue)

	if field == "" {
		return nil
	}

	condValue := parseConditionValue(rawValue)

	return &UISchemaRule{
		Effect: effect,
		Condition: &UISchemaCondition{
			Scope: "#/properties/" + field,
			Schema: &JSONSchema{
				Const: condValue,
			},
		},
	}
}

// ParseFormRuleExpression parses a rule expression from a form tag value.
// Form tags use ":" as the field:value separator (since "=" is the key=value
// delimiter in form tags). The first ":" is converted to "=" and then
// delegated to ParseRuleExpression.
func ParseFormRuleExpression(expr string, effect string) *UISchemaRule {
	normalized := strings.Replace(expr, ":", "=", 1)
	return ParseRuleExpression(normalized, effect)
}

// parseConditionValue converts a string condition value to the appropriate Go type.
// Supports: bool ("true"/"false"), integer, float, and falls back to string.
func parseConditionValue(val string) any {
	// Try bool.
	if b, err := strconv.ParseBool(val); err == nil {
		return b
	}

	// Try integer.
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i
	}

	// Try float.
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}

	return val
}
