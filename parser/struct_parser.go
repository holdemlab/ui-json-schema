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

// layoutHorizontal is the form tag value for horizontal layout grouping.
const layoutHorizontal = "horizontal"

// GenerateJSONSchema generates a JSON Schema (Draft 7) from a Go value.
// The value should be a struct or a pointer to a struct.
func GenerateJSONSchema(v any) (*schema.JSONSchema, error) {
	return GenerateJSONSchemaWithOptions(v, schema.DefaultOptions())
}

// GenerateJSONSchemaWithOptions generates a JSON Schema using the supplied options.
func GenerateJSONSchemaWithOptions(v any, opts schema.Options) (*schema.JSONSchema, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	root := &schema.JSONSchema{
		Schema: opts.DraftURL(),
		Type:   "object",
	}
	root.Properties = make(map[string]*schema.JSONSchema)

	if t.Kind() == reflect.Struct {
		parseStructFields(t, root)
	}

	return root, nil
}

// parseStructFields iterates over struct fields and populates the schema properties.
func parseStructFields(t reflect.Type, s *schema.JSONSchema) {
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

		prop := typeToSchema(field.Type)

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

	if tags.Description != "" {
		prop.Description = tags.Description
	}

	if tags.MinLength != nil {
		prop.MinLength = tags.MinLength
	}

	if tags.MaxLength != nil {
		prop.MaxLength = tags.MaxLength
	}

	if tags.Minimum != nil {
		prop.Minimum = tags.Minimum
	}

	if tags.Maximum != nil {
		prop.Maximum = tags.Maximum
	}

	if tags.Pattern != "" {
		prop.Pattern = tags.Pattern
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
func typeToSchema(t reflect.Type) *schema.JSONSchema {
	// Unwrap pointer types.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
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
		items := typeToSchema(t.Elem())
		return &schema.JSONSchema{
			Type:  "array",
			Items: items,
		}

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return &schema.JSONSchema{Type: "object"}
		}

		additional := typeToSchema(t.Elem())
		return &schema.JSONSchema{
			Type:                 "object",
			AdditionalProperties: additional,
		}

	case reflect.Struct:
		obj := &schema.JSONSchema{
			Type:       "object",
			Properties: make(map[string]*schema.JSONSchema),
		}
		parseStructFields(t, obj)
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
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	root := schema.NewVerticalLayout()

	if t.Kind() == reflect.Struct {
		buildUIElements(t, "#/properties", root, &opts)
	}

	// If any fields have categories, wrap elements into a Categorization.
	if hasCategorizedElements(root) {
		return buildCategorization(root, &opts), nil
	}

	// Apply horizontal grouping on the root layout.
	root.Elements = groupHorizontalElements(root.Elements)

	return root, nil
}

// buildUIElements iterates over struct fields and builds UI Schema elements.
func buildUIElements(t reflect.Type, basePath string, parent *schema.UISchemaElement, opts *schema.Options) {
	for i := range t.NumField() {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		name := fieldJSONName(field)
		if name == "-" {
			continue
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

			buildUIElements(fieldType, scope+"/properties", group, opts)
			// Apply horizontal grouping within nested groups immediately,
			// as groups are not affected by categorization.
			group.Elements = groupHorizontalElements(group.Elements)
			// Apply rule from the struct field tags to the Group element.
			applyRule(group, tags)
			// Propagate category, category rule & i18n from the form tag
			// so nested structs are placed into the correct Category.
			applyGroupCategoryOptions(group, formOpts)
			parent.Elements = append(parent.Elements, group)

			continue
		}

		control := buildControl(scope, name, formOpts, tags, opts)

		if formOpts.Layout == layoutHorizontal {
			ensureOptions(control)
			control.Options["layout"] = layoutHorizontal

			if formOpts.LayoutGroup != "" {
				control.Options["layoutGroup"] = formOpts.LayoutGroup
			}
		}

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
	applyCategoryRuleOptions(control, formOpts)
	applyCategoryI18nOption(control, formOpts)

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

// isHorizontalElement checks whether an element is marked with layout=horizontal.
func isHorizontalElement(el *schema.UISchemaElement) bool {
	if el.Options == nil {
		return false
	}

	v, ok := el.Options["layout"]
	return ok && v == layoutHorizontal
}

// layoutGroupName returns the named group for a horizontal element, or "".
func layoutGroupName(el *schema.UISchemaElement) string {
	if el.Options == nil {
		return ""
	}

	g, ok := el.Options["layoutGroup"].(string)
	if !ok {
		return ""
	}

	return g
}

// consumeLayoutOption removes the internal "layout" and "layoutGroup"
// options from an element after it has been consumed by the grouping logic.
func consumeLayoutOption(el *schema.UISchemaElement) {
	if el.Options == nil {
		return
	}

	delete(el.Options, "layout")
	delete(el.Options, "layoutGroup")

	if len(el.Options) == 0 {
		el.Options = nil
	}
}

// groupHorizontalElements groups elements marked with layout=horizontal
// into HorizontalLayout containers.
//
// Unnamed horizontal elements (no layoutGroup) are grouped when consecutive.
// Named horizontal elements (layoutGroup set) are collected across all
// positions and placed into a single HorizontalLayout at the position of
// the first element in that group. A single horizontal element (named or
// unnamed) is kept as a plain Control.
func groupHorizontalElements(elements []*schema.UISchemaElement) []*schema.UISchemaElement {
	// First pass: collect named groups preserving insertion order
	// and record which group each element belongs to (by pointer).
	namedGroups := make(map[string][]*schema.UISchemaElement)
	memberGroup := make(map[*schema.UISchemaElement]string)

	for _, el := range elements {
		if !isHorizontalElement(el) {
			continue
		}

		g := layoutGroupName(el)
		if g == "" {
			continue
		}

		namedGroups[g] = append(namedGroups[g], el)
		memberGroup[el] = g
	}

	// Second pass: build result, handling unnamed consecutive groups
	// and emitting named groups at first-occurrence position.
	var result []*schema.UISchemaElement
	var pendingH []*schema.UISchemaElement // unnamed consecutive buffer
	emittedGroups := make(map[string]bool)

	flushUnnamed := func() {
		if len(pendingH) == 0 {
			return
		}

		for _, el := range pendingH {
			consumeLayoutOption(el)
		}

		if len(pendingH) == 1 {
			result = append(result, pendingH[0])
		} else {
			hl := schema.NewHorizontalLayout()
			hl.Elements = append(hl.Elements, pendingH...)
			result = append(result, hl)
		}

		pendingH = nil
	}

	emitNamedGroup := func(groupName string) {
		members := namedGroups[groupName]

		for _, el := range members {
			consumeLayoutOption(el)
		}

		if len(members) == 1 {
			result = append(result, members[0])
		} else {
			hl := schema.NewHorizontalLayout()
			hl.Elements = append(hl.Elements, members...)
			result = append(result, hl)
		}

		emittedGroups[groupName] = true
	}

	for _, el := range elements {
		if g, ok := memberGroup[el]; ok {
			// Named group member: emit entire group at first occurrence,
			// skip subsequent members (already collected in pass 1).
			flushUnnamed()

			if !emittedGroups[g] {
				emitNamedGroup(g)
			}
		} else if isHorizontalElement(el) {
			// Unnamed: consecutive grouping.
			pendingH = append(pendingH, el)
		} else {
			flushUnnamed()
			result = append(result, el)
		}
	}

	flushUnnamed()

	return result
}

// applyGroupCategoryOptions propagates category, category-rule and
// category-i18n hints from the form tag of a struct field onto its
// Group element. This ensures nested structs with
// form:"category=General" are placed into the correct category
// instead of falling into "Other".
func applyGroupCategoryOptions(group *schema.UISchemaElement, formOpts schema.FormOptions) {
	if formOpts.Category != "" {
		ensureOptions(group)
		group.Options["category"] = formOpts.Category
	}

	applyCategoryRuleOptions(group, formOpts)
	applyCategoryI18nOption(group, formOpts)
}

// buildCategorization groups elements by their category option
// into a Categorization layout. Elements without a category are
// placed into an "Other" category.
func buildCategorization(root *schema.UISchemaElement, opts *schema.Options) *schema.UISchemaElement {
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
		cat := catMap[name]
		cat.Elements = groupHorizontalElements(cat.Elements)
		extractCategoryRule(cat)
		extractCategoryI18n(cat, opts)
		categorization.Elements = append(categorization.Elements, cat)
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

// applyCategoryRuleOptions stores a category-level rule hint in the control's
// Options map. The hint is later consumed by extractCategoryRule when building
// the Categorization layout. Only the first matching effect is stored
// (priority: visibleIf → hideIf → enableIf → disableIf).
func applyCategoryRuleOptions(el *schema.UISchemaElement, formOpts schema.FormOptions) {
	var effect, expr string

	switch {
	case formOpts.VisibleIf != "":
		effect, expr = schema.EffectShow, formOpts.VisibleIf
	case formOpts.HideIf != "":
		effect, expr = schema.EffectHide, formOpts.HideIf
	case formOpts.EnableIf != "":
		effect, expr = schema.EffectEnable, formOpts.EnableIf
	case formOpts.DisableIf != "":
		effect, expr = schema.EffectDisable, formOpts.DisableIf
	}

	if expr == "" {
		return
	}

	ensureOptions(el)
	el.Options["categoryRuleEffect"] = effect
	el.Options["categoryRuleExpr"] = expr
}

// extractCategoryRule scans the children of a Category element for a
// category rule hint (stored by applyCategoryRuleOptions). The first
// hint found is converted into a Rule on the Category itself and then
// removed from the child element's Options.
func extractCategoryRule(cat *schema.UISchemaElement) {
	for _, el := range cat.Elements {
		if el.Options == nil {
			continue
		}

		effect, hasEffect := el.Options["categoryRuleEffect"].(string)
		expr, hasExpr := el.Options["categoryRuleExpr"].(string)

		if !hasEffect || !hasExpr {
			continue
		}

		cat.Rule = schema.ParseFormRuleExpression(expr, effect)

		delete(el.Options, "categoryRuleEffect")
		delete(el.Options, "categoryRuleExpr")

		if len(el.Options) == 0 {
			el.Options = nil
		}

		return
	}
}

// applyCategoryI18nOption stores a category i18n key hint in the control's
// Options map. The hint is later consumed by extractCategoryI18n when building
// the Categorization layout.
func applyCategoryI18nOption(el *schema.UISchemaElement, formOpts schema.FormOptions) {
	if formOpts.I18nKey == "" {
		return
	}

	ensureOptions(el)
	el.Options["categoryI18n"] = formOpts.I18nKey
}

// extractCategoryI18n scans the children of a Category element for a
// categoryI18n hint. The first hint found sets the I18n field on the
// Category and translates the label via the Translator if available.
// The hint is then removed from the child element's Options.
func extractCategoryI18n(cat *schema.UISchemaElement, opts *schema.Options) {
	for _, el := range cat.Elements {
		if el.Options == nil {
			continue
		}

		i18nKey, ok := el.Options["categoryI18n"].(string)
		if !ok || i18nKey == "" {
			continue
		}

		cat.I18n = i18nKey
		cat.Label = translateLabel(cat.Label, i18nKey, opts)

		delete(el.Options, "categoryI18n")

		if len(el.Options) == 0 {
			el.Options = nil
		}

		return
	}
}
