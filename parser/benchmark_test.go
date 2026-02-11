package parser_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/holdemlab/ui-json-schema/parser"
)

// --- Benchmark structs ---

// BenchSmall is a small struct with 5 fields.
type BenchSmall struct {
	Name     string  `json:"name" required:"true"`
	Age      int     `json:"age" default:"25"`
	Email    string  `json:"email" format:"email"`
	Score    float64 `json:"score"`
	IsActive bool    `json:"is_active" default:"true"`
}

// BenchMedium is a medium struct with nested objects and 15+ fields.
type BenchMedium struct {
	ID       int               `json:"id" form:"hidden"`
	Name     string            `json:"name" required:"true" form:"label=Full name"`
	Email    string            `json:"email" format:"email" required:"true"`
	Age      int               `json:"age" default:"25"`
	Score    float64           `json:"score"`
	IsActive bool              `json:"is_active" default:"true"`
	Role     string            `json:"role" enum:"admin,user,moderator"`
	Bio      string            `json:"bio" form:"multiline"`
	Phone    string            `json:"phone"`
	Website  string            `json:"website" format:"uri"`
	Address  BenchAddress      `json:"address"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`
	Details  string            `json:"details" visibleIf:"is_active=true"`
	Notes    string            `json:"notes" form:"readonly"`
}

// BenchAddress represents a nested address struct.
type BenchAddress struct {
	Street  string `json:"street" required:"true"`
	City    string `json:"city" required:"true"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country" default:"US"`
}

// BenchLarge is a large struct simulating 40+ fields with nesting.
type BenchLarge struct {
	F01 string            `json:"f01" required:"true"`
	F02 string            `json:"f02" format:"email"`
	F03 int               `json:"f03" default:"10"`
	F04 float64           `json:"f04"`
	F05 bool              `json:"f05" default:"true"`
	F06 string            `json:"f06" enum:"a,b,c"`
	F07 string            `json:"f07" form:"multiline"`
	F08 string            `json:"f08" form:"readonly"`
	F09 string            `json:"f09" visibleIf:"f05=true"`
	F10 string            `json:"f10"`
	F11 string            `json:"f11"`
	F12 int               `json:"f12"`
	F13 float64           `json:"f13"`
	F14 bool              `json:"f14"`
	F15 string            `json:"f15"`
	F16 string            `json:"f16"`
	F17 int               `json:"f17"`
	F18 float64           `json:"f18"`
	F19 bool              `json:"f19"`
	F20 string            `json:"f20"`
	N1  BenchAddress      `json:"n1"`
	N2  BenchAddress      `json:"n2"`
	N3  BenchAddress      `json:"n3"`
	F21 string            `json:"f21"`
	F22 string            `json:"f22"`
	F23 int               `json:"f23"`
	F24 float64           `json:"f24"`
	F25 bool              `json:"f25"`
	F26 string            `json:"f26"`
	F27 string            `json:"f27"`
	F28 int               `json:"f28"`
	F29 float64           `json:"f29"`
	F30 bool              `json:"f30"`
	F31 []string          `json:"f31"`
	F32 []int             `json:"f32"`
	F33 map[string]string `json:"f33"`
	F34 map[string]int    `json:"f34"`
	F35 string            `json:"f35"`
	F36 string            `json:"f36"`
	F37 int               `json:"f37"`
	F38 float64           `json:"f38"`
	F39 bool              `json:"f39"`
	F40 string            `json:"f40"`
}

// --- JSON Schema from struct benchmarks ---

func BenchmarkGenerateJSONSchema_Small(b *testing.B) {
	v := BenchSmall{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateJSONSchema(v)
	}
}

func BenchmarkGenerateJSONSchema_Medium(b *testing.B) {
	v := BenchMedium{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateJSONSchema(v)
	}
}

func BenchmarkGenerateJSONSchema_Large(b *testing.B) {
	v := BenchLarge{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateJSONSchema(v)
	}
}

// --- UI Schema from struct benchmarks ---

func BenchmarkGenerateUISchema_Small(b *testing.B) {
	v := BenchSmall{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateUISchema(v)
	}
}

func BenchmarkGenerateUISchema_Medium(b *testing.B) {
	v := BenchMedium{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateUISchema(v)
	}
}

func BenchmarkGenerateUISchema_Large(b *testing.B) {
	v := BenchLarge{}
	b.ResetTimer()

	for range b.N {
		_, _ = parser.GenerateUISchema(v)
	}
}

// --- JSON parsing benchmarks ---

func BenchmarkGenerateFromJSON_Small(b *testing.B) {
	data := []byte(`{"name":"John","age":30,"score":9.5,"is_active":true,"email":"john@example.com"}`)
	b.ResetTimer()

	for range b.N {
		_, _, _ = parser.GenerateFromJSON(data)
	}
}

func BenchmarkGenerateFromJSON_Medium(b *testing.B) {
	data := buildMediumJSON()
	b.ResetTimer()

	for range b.N {
		_, _, _ = parser.GenerateFromJSON(data)
	}
}

func BenchmarkGenerateFromJSON_1MB(b *testing.B) {
	data := buildLargeJSON(1)
	b.Logf("JSON size: %d bytes (%.2f MB)", len(data), float64(len(data))/(1<<20))
	b.ResetTimer()

	for range b.N {
		_, _, _ = parser.GenerateFromJSON(data)
	}
}

func BenchmarkGenerateFromJSON_2MB(b *testing.B) {
	data := buildLargeJSON(2)
	b.Logf("JSON size: %d bytes (%.2f MB)", len(data), float64(len(data))/(1<<20))
	b.ResetTimer()

	for range b.N {
		_, _, _ = parser.GenerateFromJSON(data)
	}
}

// --- helpers ---

func buildMediumJSON() []byte {
	obj := map[string]any{
		"name":      "Alice",
		"age":       28,
		"email":     "alice@example.com",
		"is_active": true,
		"score":     9.5,
		"role":      "admin",
		"address": map[string]any{
			"street":  "123 Main St",
			"city":    "Kyiv",
			"state":   "UA",
			"zip":     "01001",
			"country": "Ukraine",
		},
		"tags":     []any{"go", "json", "schema"},
		"metadata": map[string]any{"key1": "val1", "key2": "val2"},
	}

	data, _ := json.Marshal(obj)
	return data
}

func buildLargeJSON(targetMB int) []byte {
	targetSize := targetMB * (1 << 20)

	obj := make(map[string]any)

	for i := range 50 {
		key := fmt.Sprintf("field_%03d", i)
		obj[key] = map[string]any{
			"name":     fmt.Sprintf("User %d", i),
			"email":    fmt.Sprintf("user%d@example.com", i),
			"age":      25 + i,
			"active":   i%2 == 0,
			"score":    3.14 * float64(i),
			"tags":     []any{"tag1", "tag2", "tag3"},
			"metadata": map[string]any{"k": "v"},
		}
	}

	base, _ := json.Marshal(obj)
	remaining := targetSize - len(base)

	if remaining > 0 {
		padding := make([]byte, remaining-100)
		for i := range padding {
			padding[i] = 'x'
		}

		obj["_padding"] = string(padding)
	}

	data, _ := json.Marshal(obj)
	return data
}
