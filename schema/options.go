package schema

// Options configures the behavior of JSON Schema and UI Schema generation.
type Options struct {
	// Translator is used to localize labels. Nil means no translation.
	Translator Translator
	// Locale selects the locale passed to Translator (e.g. "uk", "en").
	Locale string
	// Draft selects the JSON Schema draft version ("draft-07" or "2019-09").
	// Default is "draft-07".
	Draft string
	// Renderers maps a JSON Schema property scope to a custom renderer name.
	// The renderer name is placed into the UI Schema element's options.
	Renderers map[string]string
	// RolePermissions maps role names to field permission overrides.
	// Each permission set maps field JSON names to access levels.
	RolePermissions map[string]FieldPermissions
	// Role is the active role to apply permissions for.
	Role string
}

// FieldPermissions maps field JSON names to access levels.
type FieldPermissions map[string]AccessLevel

// AccessLevel defines the access level for a field.
type AccessLevel int

const (
	// AccessReadWrite is the default full-access level.
	AccessReadWrite AccessLevel = iota
	// AccessReadOnly makes the field readonly in UI Schema.
	AccessReadOnly
	// AccessHidden removes the field from UI Schema entirely.
	AccessHidden
)

// DraftURL returns the $schema URL for the configured draft version.
func (o Options) DraftURL() string {
	if o.Draft == "2019-09" {
		return "https://json-schema.org/draft/2019-09/schema"
	}

	return "http://json-schema.org/draft-07/schema#"
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Draft: "draft-07",
	}
}
