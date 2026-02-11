# ui-json-schema

Automatic **JSON Schema (Draft 7)** and **UI Schema (JSON Forms)** generation from Go structs and JSON objects.

## Features

- Generate JSON Schema from Go structs via `reflect`
- Generate UI Schema compatible with [JSON Forms](https://jsonforms.io/)
- Generate both schemas from arbitrary JSON objects
- **OpenAPI 3.x** → JSON Schema + UI Schema conversion
- Support for struct tags: `required`, `default`, `enum`, `format`, `form`, `i18n`, `renderer`
- UI controls: labels, hidden fields, readonly, multiline
- Conditional rules: `visibleIf`, `hideIf`, `enableIf`, `disableIf`
- **i18n** — localised labels via `Translator` interface
- **Custom renderers** — per-field renderer via tag or options map
- **Role-based permissions** — readonly / hidden fields per role
- **Categorization layouts** — tab-based UI via `form:"category=..."` tag
- **JSON Schema Draft 2019-09** support (configurable)
- HTTP API with type registry
- No external dependencies

## Installation

```bash
go get github.com/holdemlab/ui-json-schema
```

## Quick Start

### From Go structs

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/holdemlab/ui-json-schema/parser"
)

type User struct {
    ID        int    `json:"id" form:"hidden"`
    Name      string `json:"name" required:"true" form:"label=Full name"`
    Email     string `json:"email" required:"true" format:"email"`
    IsActive  bool   `json:"is_active" default:"true"`
    Role      string `json:"role" enum:"admin,user,moderator"`
    Bio       string `json:"bio" form:"multiline"`
    CreatedAt string `json:"created_at" form:"readonly"`
    Details   string `json:"details" visibleIf:"is_active=true"`
}

func main() {
    // Generate JSON Schema
    schema, _ := parser.GenerateJSONSchema(User{})

    // Generate UI Schema
    uiSchema, _ := parser.GenerateUISchema(User{})

    s, _ := json.MarshalIndent(schema, "", "  ")
    fmt.Println("JSON Schema:", string(s))

    u, _ := json.MarshalIndent(uiSchema, "", "  ")
    fmt.Println("UI Schema:", string(u))
}
```

### From JSON objects

```go
data := []byte(`{
    "name": "John",
    "age": 30,
    "is_active": true,
    "address": {
        "city": "Kyiv",
        "street": "Main St"
    }
}`)

schema, uiSchema, err := parser.GenerateFromJSON(data)
```

### HTTP API

```go
package main

import (
    "log"
    "net/http"

    handler "github.com/holdemlab/ui-json-schema/api"
)

type User struct {
    Name  string `json:"name" required:"true"`
    Email string `json:"email" format:"email"`
}

func main() {
    registry := handler.NewRegistry()
    registry.Register("User", User{})

    h := handler.NewHandler(registry)

    mux := http.NewServeMux()
    mux.HandleFunc("/schema/generate", h.GenerateHandler)

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

#### Generate from registered type

```bash
curl -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"type": "User"}'
```

#### Generate from JSON payload

```bash
curl -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"data": {"name": "John", "age": 30, "active": true}}'
```

#### Response format

```json
{
  "schema": {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {
      "name": { "type": "string" },
      "email": { "type": "string", "format": "email" }
    },
    "required": ["name"]
  },
  "uischema": {
    "type": "VerticalLayout",
    "elements": [
      { "type": "Control", "scope": "#/properties/name" },
      { "type": "Control", "scope": "#/properties/email" }
    ]
  }
}
```

## Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `json` | Field name in JSON | `json:"name"` |
| `required` | Mark field as required | `required:"true"` |
| `default` | Default value | `default:"active"` |
| `enum` | Allowed values | `enum:"a,b,c"` |
| `format` | JSON Schema format | `format:"email"` |
| `form` | UI options (label, hidden, readonly, multiline, category, layout) | `form:"label=Name;readonly"` |
| `visibleIf` | Show when condition is met | `visibleIf:"active=true"` |
| `hideIf` | Hide when condition is met | `hideIf:"role=admin"` |
| `enableIf` | Enable when condition is met | `enableIf:"agreed=true"` |
| `disableIf` | Disable when condition is met | `disableIf:"locked=true"` |
| `i18n` | Translation key for label | `i18n:"user.name"` |
| `renderer` | Custom renderer name | `renderer:"color-picker"` |

## Supported Types

| Go Type | JSON Schema Type |
|---------|-----------------|
| `string` | `string` |
| `int`, `int8`–`int64` | `integer` |
| `uint`, `uint8`–`uint64` | `integer` |
| `float32`, `float64` | `number` |
| `bool` | `boolean` |
| `time.Time` | `string` (format: `date-time`) |
| `[]T` | `array` (items: T) |
| `map[string]T` | `object` (additionalProperties: T) |
| nested `struct` | `object` (properties) |

## Project Structure

```
├── schema/
│   ├── jsonschema.go     # JSON Schema types
│   ├── uischema.go       # UI Schema types, form/rule parsing
│   ├── tags.go           # Struct tag parsing
│   ├── i18n.go           # Translator interface & MapTranslator
│   └── options.go        # Generation options (draft, i18n, renderers, roles)
├── parser/
│   ├── struct_parser.go  # Schema generation from Go structs
│   ├── json_parser.go    # Schema generation from JSON objects
│   └── openapi_parser.go # OpenAPI 3.x → JSON Schema + UI Schema
├── api/
│   ├── registry.go       # Type registry
│   └── handler.go        # HTTP handler
└── cmd/server/
    └── main.go           # Server entry point
```

## Advanced Usage

### i18n (Localised Labels)

```go
tr := schema.NewMapTranslator(map[string]map[string]string{
    "uk": {
        "user.name":  "Ім'я",
        "user.email": "Електронна пошта",
    },
})

opts := schema.Options{
    Translator: tr,
    Locale:     "uk",
}

type User struct {
    Name  string `json:"name" i18n:"user.name" form:"label=Name"`
    Email string `json:"email" i18n:"user.email"`
}

ui, _ := parser.GenerateUISchemaWithOptions(User{}, opts)
// Name label → "Ім'я", Email label → "Електронна пошта"
```

### Custom Renderers

```go
// Via struct tag:
type Config struct {
    Color string `json:"color" renderer:"color-picker"`
}

// Or via options map:
opts := schema.Options{
    Renderers: map[string]string{
        "#/properties/rating": "star-rating",
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Config{}, opts)
```

### Role-Based Permissions

```go
opts := schema.Options{
    Role: "viewer",
    RolePermissions: map[string]schema.FieldPermissions{
        "viewer": {
            "name":  schema.AccessReadOnly,
            "role":  schema.AccessHidden,
        },
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(User{}, opts)
// "name" is readonly, "role" is excluded from UI Schema
```

### Categorization (Tab Layout)

```go
type Form struct {
    Name  string `json:"name" form:"category=Personal"`
    Email string `json:"email" form:"category=Personal"`
    Role  string `json:"role" form:"category=Work"`
    Bio   string `json:"bio" form:"category=Work;multiline"`
}

ui, _ := parser.GenerateUISchemaWithOptions(Form{}, schema.DefaultOptions())
// Root type is "Categorization" with Category children "Personal" and "Work"
```
```json
{
  "type": "Categorization",
  "elements": [
    {
      "type": "Category",
      "label": "Personal",
      "elements": [
        {
          "type": "Control",
          "scope": "#/properties/name"
        },
        {
          "type": "Control",
          "scope": "#/properties/email"
        }
      ]
    },
    {
      "type": "Category",
      "label": "Work",
      "elements": [
        {
          "type": "Control",
          "scope": "#/properties/role"
        },
        {
          "type": "Control",
          "scope": "#/properties/bio",
          "options": {
            "multi": true
          }
        }
      ]
    }
  ]
}
```

### JSON Schema Draft 2019-09

```go
opts := schema.Options{Draft: "2019-09"}
s, _ := parser.GenerateJSONSchemaWithOptions(User{}, opts)
// $schema → "https://json-schema.org/draft/2019-09/schema"
```

### OpenAPI 3.x → JSON Forms

```go
openAPIDoc := []byte(`{
    "components": {
        "schemas": {
            "User": {
                "type": "object",
                "properties": {
                    "name": {"type": "string"},
                    "email": {"type": "string", "format": "email"}
                },
                "required": ["name"]
            }
        }
    }
}`)

schema, uiSchema, _ := parser.GenerateFromOpenAPI(openAPIDoc, "User")
```

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-cover

# Run benchmarks
make bench

# Run linter
make lint

# Build
make build

# Run server
make run
```

## Performance

Benchmarks on Intel i7-14700HX:

| Operation | Time | Allocations |
|-----------|------|-------------|
| JSON Schema from small struct (5 fields) | ~2.2 µs | 9 allocs |
| JSON Schema from medium struct (15 fields) | ~10.4 µs | 42 allocs |
| JSON Schema from large struct (40+ fields) | ~25.5 µs | 93 allocs |
| Generate from 1 MB JSON | ~3.9 ms | 3,536 allocs |
| Generate from 2 MB JSON | ~6.3 ms | 3,536 allocs |

All operations are well under the 100 ms target for JSON up to 2 MB.

## License

MIT
