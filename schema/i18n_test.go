package schema_test

import (
	"testing"

	"github.com/holdemlab/ui-json-schema/schema"
)

func TestMapTranslator_Translate(t *testing.T) {
	tr := schema.NewMapTranslator(map[string]map[string]string{
		"uk": {
			"name":  "Ім'я",
			"email": "Електронна пошта",
		},
		"en": {
			"name":  "Name",
			"email": "Email",
		},
	})

	tests := []struct {
		name     string
		key      string
		locale   string
		expected string
	}{
		{"uk name", "name", "uk", "Ім'я"},
		{"uk email", "email", "uk", "Електронна пошта"},
		{"en name", "name", "en", "Name"},
		{"en email", "email", "en", "Email"},
		{"missing key", "phone", "uk", "phone"},
		{"missing locale", "name", "de", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tr.Translate(tt.key, tt.locale)
			if result != tt.expected {
				t.Errorf("Translate(%q, %q) = %q, want %q", tt.key, tt.locale, result, tt.expected)
			}
		})
	}
}

func TestMapTranslator_NilMap(t *testing.T) {
	tr := schema.NewMapTranslator(nil)

	result := tr.Translate("name", "uk")
	if result != "name" {
		t.Errorf("expected fallback to key %q, got %q", "name", result)
	}
}

func TestOptions_DraftURL(t *testing.T) {
	tests := []struct {
		name     string
		draft    string
		expected string
	}{
		{"default", "", "http://json-schema.org/draft-07/schema#"},
		{"draft-07", "draft-07", "http://json-schema.org/draft-07/schema#"},
		{"2019-09", "2019-09", "https://json-schema.org/draft/2019-09/schema"},
		{"unknown", "future", "http://json-schema.org/draft-07/schema#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := schema.Options{Draft: tt.draft}
			if got := opts.DraftURL(); got != tt.expected {
				t.Errorf("DraftURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := schema.DefaultOptions()
	if opts.Draft != "draft-07" {
		t.Errorf("expected default draft 'draft-07', got %q", opts.Draft)
	}

	if opts.DraftURL() != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("expected draft-07 URL, got %q", opts.DraftURL())
	}
}
