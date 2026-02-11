// Package parser provides functions for parsing Go structs
// and raw JSON objects into intermediate representations
// used by the schema generator.
package parser

import (
	"reflect"
	"strings"
	"time"

	"github.com/holdemlab/ui-json-schema/schema"
)

// timeType is cached to avoid repeated reflect.TypeOf calls.
var timeType = reflect.TypeOf(time.Time{})

// GenerateJSONSchema generates a JSON Schema (Draft 7) from a Go value.
// The value should be a struct or a pointer to a struct.
func GenerateJSONSchema(v any) (*schema.JSONSchema, error) {
	return GenerateJSONSchemaWithOptions(v, schema.DefaultOptions())
}

// GenerateJSONSchemaWithOptions generates a JSON Schema using the supplied options.
func GenerateJSONSchemaWithOptions(v any, opts schema.Options) (*schema.JSONSchema, error) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	root := &schema.JSONSchema{
		Schema: opts.DraftURL(),
		Type:   "object",
	}
	root.Properties = make(map[string]*schema.JSONSchema)

	if t.Kind() == reflect.Struct {
		parseStructFields(t, val, root, &opts)
	}

	return root, nil
}

// parseStructFields iterates over struct fields and populates the schema properties.
func parseStructFields(t reflect.Type, val reflect.Value, s *schema.JSONSchema, opts *schema.Options) {
	for i := range t.NumField() {
		field := t.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		name := fieldJSONName(field)
		if name == "-" {
			continue
		}

		// When OmitEmpty is enabled, skip fields tagged with omitempty
		// whose value is the zero value for their type.
		if opts != nil && opts.OmitEmpty && fieldHasOmitempty(field) {
			if val.IsValid() && isZeroValue(val.Field(i)) {
				continue
			}
		}

		var fieldVal reflect.Value
		if val.IsValid() {
			fieldVal = val.Field(i)
		}

		prop := typeToSchema(field.Type, fieldVal, opts)

		// Apply struct tags to the property.
		tags := schema.ParseFieldTags(field)
		applyTags(prop, tags)

		// Add to required list if tagged.
		if tags.Required {
			s.Required = append(s.Required, name)
		}

		s.Properties[name] = prop
	}
}

// applyTags applies parsed struct tag values to a JSON Schema property.
func applyTags(prop *schema.JSONSchema, tags schema.FieldTags) {
	if tags.Default != nil {
		prop.Default = tags.Default
	}

	if len(tags.Enum) > 0 {
		prop.Enum = tags.Enum
	}

	if tags.Format != "" {
		prop.Format = tags.Format
	}
}

// fieldHasOmitempty checks whether the field's json tag contains "omitempty".
func fieldHasOmitempty(field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	_, opts, _ := strings.Cut(tag, ",")

	for opts != "" {
		var name string
		name, opts, _ = strings.Cut(opts, ",")

		if name == "omitempty" {
			return true
		}
	}

	return false
}

// isZeroValue reports whether v is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() { //nolint:exhaustive // covers JSON-representable types
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

// fieldJSONName returns the JSON field name from the json struct tag.
// Falls back to the Go field name if no tag is present.
func fieldJSONName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name
	}

	return name
}

// typeToSchema converts a reflect.Type to a JSONSchema property.
func typeToSchema(t reflect.Type, val reflect.Value, opts *schema.Options) *schema.JSONSchema {
	// Unwrap pointer types.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()

		if val.IsValid() && !val.IsNil() {
			val = val.Elem()
		} else {
			val = reflect.Value{}
		}
	}

	// Handle time.Time as a special case.
	if t == timeType {
		return &schema.JSONSchema{
			Type:   "string",
			Format: "date-time",
		}
	}

	switch t.Kind() { //nolint:exhaustive // only JSON-representable types are handled
	case reflect.String:
		return &schema.JSONSchema{Type: "string"}

	case reflect.Bool:
		return &schema.JSONSchema{Type: "boolean"}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &schema.JSONSchema{Type: "integer"}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &schema.JSONSchema{Type: "integer"}

	case reflect.Float32, reflect.Float64:
		return &schema.JSONSchema{Type: "number"}

	case reflect.Slice, reflect.Array:
		items := typeToSchema(t.Elem(), reflect.Value{}, opts)
		return &schema.JSONSchema{
			Type:  "array",
			Items: items,
		}

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return &schema.JSONSchema{Type: "object"}
		}

		additional := typeToSchema(t.Elem(), reflect.Value{}, opts)
		return &schema.JSONSchema{
			Type:                 "object",
			AdditionalProperties: additional,
		}

	case reflect.Struct:
		obj := &schema.JSONSchema{
			Type:       "object",
			Properties: make(map[string]*schema.JSONSchema),
		}
		parseStructFields(t, val, obj, opts)
		return obj

	default:
		return &schema.JSONSchema{Type: "string"}
	}
}

// GenerateUISchema generates a JSON Forms UI Schema from a Go value.
// The value should be a struct or a pointer to a struct.
func GenerateUISchema(v any) (*schema.UISchemaElement, error) {
	return GenerateUISchemaWithOptions(v, schema.DefaultOptions())
}

// GenerateUISchemaWithOptions generates a JSON Forms UI Schema using the supplied options.
func GenerateUISchemaWithOptions(v any, opts schema.Options) (*schema.UISchemaElement, error) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	root := schema.NewVerticalLayout()

	if t.Kind() == reflect.Struct {
		buildUIElements(t, val, "#/properties", root, &opts)
	}

	// If any fields have categories, wrap elements into a Categorization.
	if hasCategorizedElements(root) {
		return buildCategorization(root), nil
	}

	return root, nil
}

// buildUIElements iterates over struct fields and builds UI Schema elements.
func buildUIElements(t reflect.Type, val reflect.Value, basePath string, parent *schema.UISchemaElement, opts *schema.Options) {
	for i := range t.NumField() {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		name := fieldJSONName(field)
		if name == "-" {
			continue
		}

		// When OmitEmpty is enabled, skip fields tagged with omitempty
		// whose value is the zero value for their type.
		if opts != nil && opts.OmitEmpty && fieldHasOmitempty(field) {
			if val.IsValid() && isZeroValue(val.Field(i)) {
				continue
			}
		}

		tags := schema.ParseFieldTags(field)
		formOpts := schema.ParseFormTag(tags.Form)

		if isFieldHidden(name, formOpts, opts) {
			continue
		}

		scope := basePath + "/" + name
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Nested structs (excluding time.Time) get a Group layout.
		if fieldType.Kind() == reflect.Struct && fieldType != timeType {
			label := formOpts.Label
			if label == "" {
				label = field.Name
			}

			label = translateLabel(label, tags.I18nKey, opts)

			group := schema.NewGroup(label)

			var nestedVal reflect.Value
			if val.IsValid() {
				nestedVal = val.Field(i)
			}

			buildUIElements(fieldType, nestedVal, scope+"/properties", group, opts)
			parent.Elements = append(parent.Elements, group)

			continue
		}

		control := buildControl(scope, name, formOpts, tags, opts)
		parent.Elements = append(parent.Elements, control)
	}
}

// isFieldHidden determines whether a field should be excluded from the UI Schema.
func isFieldHidden(name string, formOpts schema.FormOptions, opts *schema.Options) bool {
	if formOpts.Hidden {
		return true
	}

	if opts != nil && opts.Role != "" {
		if perms, ok := opts.RolePermissions[opts.Role]; ok {
			if level, exists := perms[name]; exists && level == schema.AccessHidden {
				return true
			}
		}
	}

	return false
}

// buildControl creates a fully configured Control UI Schema element.
func buildControl(scope, name string, formOpts schema.FormOptions, tags schema.FieldTags, opts *schema.Options) *schema.UISchemaElement {
	control := schema.NewControl(scope)

	controlLabel := translateLabel(formOpts.Label, tags.I18nKey, opts)
	if controlLabel != "" {
		control.Label = controlLabel
	}

	if formOpts.Category != "" {
		ensureOptions(control)
		control.Options["category"] = formOpts.Category
	}

	isReadonly := formOpts.Readonly || isRoleReadOnly(name, opts)

	if isReadonly || formOpts.Multiline {
		ensureOptions(control)

		if isReadonly {
			control.Options["readonly"] = true
		}

		if formOpts.Multiline {
			control.Options["multi"] = true
		}
	}

	renderer := resolveRenderer(scope, tags.Renderer, opts)
	if renderer != "" {
		ensureOptions(control)
		control.Options["renderer"] = renderer
	}

	applyRule(control, tags)

	return control
}

// ensureOptions initializes the Options map on a control if it is nil.
func ensureOptions(el *schema.UISchemaElement) {
	if el.Options == nil {
		el.Options = make(map[string]any)
	}
}

// isRoleReadOnly checks whether the active role requires a field to be readonly.
func isRoleReadOnly(name string, opts *schema.Options) bool {
	if opts == nil || opts.Role == "" {
		return false
	}

	if perms, ok := opts.RolePermissions[opts.Role]; ok {
		if level, exists := perms[name]; exists && level == schema.AccessReadOnly {
			return true
		}
	}

	return false
}

// resolveRenderer determines the renderer name from a tag or from options.
func resolveRenderer(scope, tagRenderer string, opts *schema.Options) string {
	if tagRenderer != "" {
		return tagRenderer
	}

	if opts != nil && opts.Renderers != nil {
		return opts.Renderers[scope]
	}

	return ""
}

// translateLabel applies i18n translation to a label.
// If an i18n key is set and a translator is available, the key is translated.
// If the key is empty but a label exists, the label is used as the key.
func translateLabel(label, i18nKey string, opts *schema.Options) string {
	if opts == nil || opts.Translator == nil || opts.Locale == "" {
		if i18nKey != "" && label == "" {
			return i18nKey
		}

		return label
	}

	key := i18nKey
	if key == "" && label != "" {
		key = label
	}

	if key == "" {
		return ""
	}

	return opts.Translator.Translate(key, opts.Locale)
}

// hasCategorizedElements checks if any element has a category option set.
func hasCategorizedElements(root *schema.UISchemaElement) bool {
	for _, el := range root.Elements {
		if el.Options != nil {
			if _, ok := el.Options["category"]; ok {
				return true
			}
		}
	}

	return false
}

// buildCategorization groups elements by their category option
// into a Categorization layout. Elements without a category are
// placed into an "Other" category.
func buildCategorization(root *schema.UISchemaElement) *schema.UISchemaElement {
	catMap := make(map[string]*schema.UISchemaElement)
	catOrder := make([]string, 0)

	for _, el := range root.Elements {
		catName := "Other"

		if el.Options != nil {
			if c, ok := el.Options["category"].(string); ok && c != "" {
				catName = c
				// Remove the category option now that it's consumed.
				delete(el.Options, "category")

				if len(el.Options) == 0 {
					el.Options = nil
				}
			}
		}

		if _, exists := catMap[catName]; !exists {
			catMap[catName] = schema.NewCategory(catName)
			catOrder = append(catOrder, catName)
		}

		catMap[catName].Elements = append(catMap[catName].Elements, el)
	}

	categorization := schema.NewCategorization()
	for _, name := range catOrder {
		categorization.Elements = append(categorization.Elements, catMap[name])
	}

	return categorization
}

// applyRule sets the first matching rule on a control element.
// Priority order: visibleIf → hideIf → enableIf → disableIf.
func applyRule(control *schema.UISchemaElement, tags schema.FieldTags) {
	switch {
	case tags.VisibleIf != "":
		control.Rule = schema.ParseRuleExpression(tags.VisibleIf, schema.EffectShow)
	case tags.HideIf != "":
		control.Rule = schema.ParseRuleExpression(tags.HideIf, schema.EffectHide)
	case tags.EnableIf != "":
		control.Rule = schema.ParseRuleExpression(tags.EnableIf, schema.EffectEnable)
	case tags.DisableIf != "":
		control.Rule = schema.ParseRuleExpression(tags.DisableIf, schema.EffectDisable)
	}
}
