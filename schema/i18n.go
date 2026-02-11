package schema

// Translator resolves localized strings by key and locale.
// Implementations may load translations from files, databases or
// embedded maps.
type Translator interface {
	// Translate returns the localized string for the given key and locale.
	// If no translation is found, it should return the key unchanged.
	Translate(key, locale string) string
}

// MapTranslator is a simple in-memory Translator backed by nested maps.
// The outer map key is the locale (e.g. "uk", "en"), the inner key is
// the translation key, and the value is the translated string.
type MapTranslator struct {
	translations map[string]map[string]string
}

// NewMapTranslator creates a MapTranslator from a locale → key → value map.
func NewMapTranslator(m map[string]map[string]string) *MapTranslator {
	if m == nil {
		m = make(map[string]map[string]string)
	}

	return &MapTranslator{translations: m}
}

// Translate returns the translation for the given key and locale.
// Falls back to the key itself when no translation exists.
func (t *MapTranslator) Translate(key, locale string) string {
	if msgs, ok := t.translations[locale]; ok {
		if val, ok := msgs[key]; ok {
			return val
		}
	}

	return key
}
