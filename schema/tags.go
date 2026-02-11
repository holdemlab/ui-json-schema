package schema

import (
	"reflect"
	"strconv"
	"strings"
)

// FieldTags holds parsed struct tag metadata for a single field.
type FieldTags struct {
	Required bool
	Default  any
	Enum     []any
	Format   string
	// Form holds raw UI-related metadata (used in UI Schema generation).
	Form string
	// I18nKey holds the i18n translation key for the field label.
	I18nKey string
	// VisibleIf holds a SHOW condition expression like "field=value".
	VisibleIf string
	// HideIf holds a HIDE condition expression like "field=value".
	HideIf string
	// EnableIf holds an ENABLE condition expression like "field=value".
	EnableIf string
	// DisableIf holds a DISABLE condition expression like "field=value".
	DisableIf string
	// Renderer holds a custom renderer name for the field.
	Renderer string
}

// ParseFieldTags extracts schema-relevant tags from a struct field.
func ParseFieldTags(field reflect.StructField) FieldTags {
	var ft FieldTags

	if v := field.Tag.Get("required"); strings.EqualFold(v, "true") {
		ft.Required = true
	}

	if v := field.Tag.Get("default"); v != "" {
		ft.Default = parseDefaultValue(v, field.Type)
	}

	if v := field.Tag.Get("enum"); v != "" {
		ft.Enum = parseEnumValues(v)
	}

	if v := field.Tag.Get("format"); v != "" {
		ft.Format = v
	}

	if v := field.Tag.Get("form"); v != "" {
		ft.Form = v
	}

	if v := field.Tag.Get("i18n"); v != "" {
		ft.I18nKey = v
	}

	if v := field.Tag.Get("renderer"); v != "" {
		ft.Renderer = v
	}

	if v := field.Tag.Get("visibleIf"); v != "" {
		ft.VisibleIf = v
	}

	if v := field.Tag.Get("hideIf"); v != "" {
		ft.HideIf = v
	}

	if v := field.Tag.Get("enableIf"); v != "" {
		ft.EnableIf = v
	}

	if v := field.Tag.Get("disableIf"); v != "" {
		ft.DisableIf = v
	}

	return ft
}

// parseDefaultValue converts a string default value to the appropriate Go type
// based on the field's reflect.Type.
func parseDefaultValue(val string, t reflect.Type) any {
	// Unwrap pointer.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() { //nolint:exhaustive // only JSON-representable defaults
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return val
		}
		return b

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return val
		}
		return i

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return val
		}
		return u

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return val
		}
		return f

	default:
		return val
	}
}

// parseEnumValues splits a comma-separated enum string into a slice of any.
func parseEnumValues(val string) []any {
	parts := strings.Split(val, ",")
	result := make([]any, 0, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
