# ui-json-schema — Full Documentation

> A library for automatic generation of **JSON Schema (Draft 7 / 2019-09)** and **UI Schema (JSON Forms)** from Go structs, JSON objects, and OpenAPI 3.x documents.

**Module:** `github.com/holdemlab/ui-json-schema`  
**Go:** 1.24+  
**Dependencies:** none (stdlib only)

---

## Table of Contents

1. [Installation](#installation)
2. [Architecture](#architecture)
3. [`schema` Package — Types and Configuration](#schema-package)
   - [JSONSchema](#jsonschema)
   - [UISchemaElement](#uischemaelement)
   - [UISchemaRule and UISchemaCondition](#uischemarule-and-uischemacondition)
   - [FormOptions](#formoptions)
   - [FieldTags](#fieldtags)
   - [Options](#options)
   - [AccessLevel and FieldPermissions](#accesslevel-and-fieldpermissions)
   - [Translator and MapTranslator](#translator-and-maptranslator)
4. [`parser` Package — Schema Generation](#parser-package)
   - [Generation from Go Structs](#generation-from-go-structs)
   - [Generation from JSON Objects](#generation-from-json-objects)
   - [Generation from OpenAPI 3.x](#generation-from-openapi-3x)
5. [`api` Package — HTTP API](#api-package)
   - [Registry](#registry)
   - [Handler](#handler)
   - [HTTP Endpoint](#http-endpoint)
6. [Struct Tags — Complete Reference](#struct-tags)
   - [Base Tags](#base-tags)
   - [`form` Tag — UI Options](#form-tag)
   - [Conditional Rule Tags](#conditional-rule-tags)
   - [i18n and renderer](#i18n-and-renderer)
7. [Go → JSON Schema Type Mapping](#type-mapping)
8. [Usage Examples](#usage-examples)
   - [Basic Generation](#basic-generation)
   - [i18n — Localization](#i18n--localization)
   - [Custom Renderers](#custom-renderers)
   - [Roles and Permissions](#roles-and-permissions)
   - [Categorization (Tabs)](#categorization-tabs)
   - [Nested Structs in Categories](#nested-structs-in-categories)
   - [Conditional Visibility](#conditional-visibility)
   - [Validation Constraints](#validation-constraints)
   - [HorizontalLayout](#horizontallayout)
   - [Named Layout Groups](#named-layout-groups)
   - [Rules on Category](#rules-on-category)
   - [i18n on Category](#i18n-on-category)
   - [Rules on Group (Nested Structs)](#rules-on-group-nested-structs)
   - [Array Detail (Slice of Structs)](#array-detail-slice-of-structs)
   - [JSON Schema Draft 2019-09](#json-schema-draft-2019-09)
   - [OmitEmpty — Empty Field Filtering](#omitempty--empty-field-filtering)
   - [OpenAPI 3.x → JSON Forms](#openapi-3x--json-forms)
   - [Generation from JSON Data](#generation-from-json-data)
   - [HTTP Server with Type Registry](#http-server-with-type-registry)
   - [Combined Example](#combined-example)
9. [HTTP API — Reference](#http-api--reference)
10. [Performance](#performance)
11. [Development](#development)

---

## Installation

```bash
go get github.com/holdemlab/ui-json-schema
```

Package imports:

```go
import (
    "github.com/holdemlab/ui-json-schema/schema"  // types, options, i18n
    "github.com/holdemlab/ui-json-schema/parser"  // schema generation
    "github.com/holdemlab/ui-json-schema/api"     // HTTP API, registry
)
```

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  cmd/server/main.go                 │
│           HTTP server (port from $ADDR or :8080)    │
└──────────────────────┬──────────────────────────────┘
                       │ uses
┌──────────────────────▼──────────────────────────────┐
│                    api/                              │
│  handler.go  — POST /schema/generate                │
│  registry.go — thread-safe type registry            │
└──────────────────────┬──────────────────────────────┘
                       │ calls
┌──────────────────────▼──────────────────────────────┐
│                   parser/                           │
│  struct_parser.go  — Go struct → Schema             │
│  json_parser.go    — []byte JSON → Schema           │
│  openapi_parser.go — OpenAPI 3.x → Schema           │
└──────────────────────┬──────────────────────────────┘
                       │ uses
┌──────────────────────▼──────────────────────────────┐
│                   schema/                           │
│  jsonschema.go — JSONSchema type                    │
│  uischema.go   — UISchemaElement, rules, forms      │
│  tags.go       — FieldTags, struct tag parsing       │
│  i18n.go       — Translator, MapTranslator          │
│  options.go    — Options, AccessLevel, Permissions  │
└─────────────────────────────────────────────────────┘
```

The library consists of three main layers:

1. **`schema`** — defines data types (JSON Schema, UI Schema), configuration, and helper interfaces.
2. **`parser`** — contains schema generation logic from various sources (Go struct, JSON, OpenAPI).
3. **`api`** — provides an HTTP handler and type registry for use as a microservice.

---

## `schema` Package

### JSONSchema

Represents a JSON Schema document (compatible with Draft 7 and Draft 2019-09).

```go
type JSONSchema struct {
    Schema               string                 `json:"$schema,omitempty"`
    Type                 string                 `json:"type,omitempty"`
    Properties           map[string]*JSONSchema `json:"properties,omitempty"`
    Items                *JSONSchema            `json:"items,omitempty"`
    AdditionalProperties *JSONSchema            `json:"additionalProperties,omitempty"`
    Required             []string               `json:"required,omitempty"`
    Format               string                 `json:"format,omitempty"`
    Default              any                    `json:"default,omitempty"`
    Enum                 []any                  `json:"enum,omitempty"`
    Description          string                 `json:"description,omitempty"`
    Title                string                 `json:"title,omitempty"`
    Const                any                    `json:"const,omitempty"`
    MinLength            *int                   `json:"minLength,omitempty"`
    MaxLength            *int                   `json:"maxLength,omitempty"`
    Minimum              *float64               `json:"minimum,omitempty"`
    Maximum              *float64               `json:"maximum,omitempty"`
    Pattern              string                 `json:"pattern,omitempty"`
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Schema` | `string` | JSON Schema standard URL (`$schema`) |
| `Type` | `string` | Value type: `"object"`, `"string"`, `"integer"`, `"number"`, `"boolean"`, `"array"`, `"null"` |
| `Properties` | `map[string]*JSONSchema` | Object properties (for `type: "object"`) |
| `Items` | `*JSONSchema` | Array item schema (for `type: "array"`) |
| `AdditionalProperties` | `*JSONSchema` | Additional properties schema (for `map[string]T`) |
| `Required` | `[]string` | List of required fields |
| `Format` | `string` | Value format: `"email"`, `"date-time"`, `"uri"`, etc. |
| `Default` | `any` | Default value |
| `Enum` | `[]any` | List of allowed values |
| `Description` | `string` | Field description |
| `Title` | `string` | Field title |
| `Const` | `any` | Fixed value |
| `MinLength` | `*int` | Minimum string length |
| `MaxLength` | `*int` | Maximum string length |
| `Minimum` | `*float64` | Minimum numeric value |
| `Maximum` | `*float64` | Maximum numeric value |
| `Pattern` | `string` | Regex pattern for string fields |

**Constructor:**

```go
func NewJSONSchema() *JSONSchema
```

Creates a root JSON Schema object with `$schema` set to the Draft 7 URL and `type` set to `"object"`.

---

### UISchemaElement

Represents a UI Schema element for [JSON Forms](https://jsonforms.io/). Can be a layout or a control.

```go
type UISchemaElement struct {
    Type     string             `json:"type"`
    Label    string             `json:"label,omitempty"`
    I18n     string             `json:"i18n,omitempty"`
    Scope    string             `json:"scope,omitempty"`
    Elements []*UISchemaElement `json:"elements,omitempty"`
    Options  map[string]any     `json:"options,omitempty"`
    Rule     *UISchemaRule      `json:"rule,omitempty"`
}
```

**Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `string` | Element type: `"VerticalLayout"`, `"HorizontalLayout"`, `"Group"`, `"Categorization"`, `"Category"`, `"Control"` |
| `Label` | `string` | Label (for `Group`, `Category`, `Control`) |
| `I18n` | `string` | i18n translation key (for `Category`) |
| `Scope` | `string` | JSON Pointer path to the property (for `Control`), e.g. `"#/properties/name"` |
| `Elements` | `[]*UISchemaElement` | Child elements (for layouts) |
| `Options` | `map[string]any` | Additional options (`readonly`, `multi`, `renderer`) |
| `Rule` | `*UISchemaRule` | Conditional visibility/accessibility rule |

**Layout types:**

| Type | Description | Constructor |
|------|-------------|-------------|
| `VerticalLayout` | Vertical element arrangement (default) | `NewVerticalLayout()` |
| `HorizontalLayout` | Horizontal arrangement | `NewHorizontalLayout()` |
| `Group` | Group with a heading | `NewGroup(label string)` |
| `Categorization` | Root container for tabs | `NewCategorization()` |
| `Category` | Tab (child element of `Categorization`) | `NewCategory(label string)` |

**Control:**

```go
func NewControl(scope string) *UISchemaElement
```

Creates a control that points to a JSON Schema property via `scope`.

---

### UISchemaRule and UISchemaCondition

Rules allow dynamically showing/hiding/enabling/disabling UI elements based on other field values.

```go
type UISchemaRule struct {
    Effect    string             `json:"effect"`
    Condition *UISchemaCondition `json:"condition"`
}

type UISchemaCondition struct {
    Scope  string      `json:"scope"`
    Schema *JSONSchema `json:"schema"`
}
```

**Effects (constants):**

| Constant | Value | Description |
|----------|-------|-------------|
| `EffectShow` | `"SHOW"` | Show element when condition is met |
| `EffectHide` | `"HIDE"` | Hide element |
| `EffectEnable` | `"ENABLE"` | Enable element |
| `EffectDisable` | `"DISABLE"` | Disable element |

**Rule parsing:**

```go
func ParseRuleExpression(expr string, effect string) *UISchemaRule
```

Accepts an expression in `"field=value"` format and an effect, returns `*UISchemaRule`. The value is automatically coerced to `bool`, `int`, `float64`, or `string`.

**Example:**

```go
rule := schema.ParseRuleExpression("is_active=true", schema.EffectShow)
// rule.Effect = "SHOW"
// rule.Condition.Scope = "#/properties/is_active"
// rule.Condition.Schema.Const = true (bool)
```

**Form tag rule parsing:**

```go
func ParseFormRuleExpression(expr string, effect string) *UISchemaRule
```

Similar to `ParseRuleExpression`, but accepts expressions in `"field:value"` format (with `:` as separator instead of `=`). Used for parsing rules from `form` tag directives (`visibleIf=field:value`).

**Example:**

```go
rule := schema.ParseFormRuleExpression("provideAddress:true", schema.EffectShow)
// rule.Effect = "SHOW"
// rule.Condition.Scope = "#/properties/provideAddress"
// rule.Condition.Schema.Const = true (bool)
```

---

### FormOptions

Stores parsed data from the `form` tag.

```go
type FormOptions struct {
    Label     string // Control label
    Hidden    bool   // Hide from UI
    Readonly  bool   // Read-only field
    Multiline bool   // Multi-line text field
    Category  string // Category (tab) name
    Layout      string // Layout override ("horizontal")
    LayoutGroup string // Named group for combining non-adjacent fields
    VisibleIf   string // SHOW rule for Category/Group ("field:value")
    HideIf    string // HIDE rule for Category/Group
    EnableIf  string // ENABLE rule for Category/Group
    DisableIf string // DISABLE rule for Category/Group
    I18nKey   string // i18n key for Category label
}
```

**Parsing:**

```go
func ParseFormTag(tag string) FormOptions
```

Parses the `form` tag value. Directives are separated by `;`.

**Tag examples:**

```
form:"label=Full Name"
form:"hidden"
form:"readonly"
form:"multiline"
form:"category=Personal Data"
form:"label=Name;readonly;category=Profile"
form:"layout=horizontal"
form:"layout=horizontal:contact"
form:"category=Address;visibleIf=provideAddress:true"
form:"category=Personal;i18n=category.personal"
```

---

### FieldTags

Stores all parsed struct tags for a single field.

```go
type FieldTags struct {
    Required  bool
    Default   any
    Enum      []any
    Format    string
    Form      string // raw form tag string
    I18nKey   string // translation key
    VisibleIf string // "field=value" for SHOW
    HideIf    string // "field=value" for HIDE
    EnableIf  string // "field=value" for ENABLE
    DisableIf string // "field=value" for DISABLE
    Renderer  string // custom renderer name
    Description string // field description
    MinLength   *int   // minimum string length
    MaxLength   *int   // maximum string length
    Minimum     *float64 // minimum numeric value
    Maximum     *float64 // maximum numeric value
    Pattern     string   // regex pattern for string fields
}
```

**Parsing:**

```go
func ParseFieldTags(field reflect.StructField) FieldTags
```

Extracts all schema-relevant tags from a field. The `default` value is automatically coerced to the field type (`bool`, `int`, `uint`, `float`, `string`). The `enum` value is split by commas.

---

### Options

Configures the behavior of schema generation.

```go
type Options struct {
    Translator      Translator            // localization interface
    Locale          string                // locale: "uk", "en", etc.
    Draft           string                // "draft-07" (default) or "2019-09"
    Renderers       map[string]string     // scope → renderer name
    RolePermissions map[string]FieldPermissions // role → field permissions
    Role            string                // active role
    OmitEmpty       bool                  // exclude omitempty fields with zero values
}
```

**Methods:**

```go
// DraftURL returns the $schema URL for the selected Draft
func (o Options) DraftURL() string

// DefaultOptions returns Options with Draft = "draft-07"
func DefaultOptions() Options
```

**`Draft` values:**

| Value | `$schema` URL |
|-------|---------------|
| `"draft-07"` (default) | `http://json-schema.org/draft-07/schema#` |
| `"2019-09"` | `https://json-schema.org/draft/2019-09/schema` |

**`OmitEmpty` field:**

When `OmitEmpty: true`, fields tagged with `json:",omitempty"` are excluded from the generated JSON Schema and UI Schema if the corresponding value is the zero value for its type. Defaults to `false` — all fields are always included. Works recursively for nested structs.

---

### AccessLevel and FieldPermissions

Access levels for role-based field management.

```go
type AccessLevel int

const (
    AccessFull     AccessLevel = iota // 0 — full access (default)
    AccessReadOnly                    // 1 — read only
    AccessHidden                      // 2 — field hidden
)

type FieldPermissions map[string]AccessLevel
```

**Usage:**

```go
opts := schema.Options{
    Role: "viewer",
    RolePermissions: map[string]schema.FieldPermissions{
        "viewer": {
            "name":  schema.AccessReadOnly, // readonly in UI
            "role":  schema.AccessHidden,   // hidden from UI
        },
        "admin": {
            // all fields accessible
        },
    },
}
```

---

### Translator and MapTranslator

Interface for label localization.

```go
type Translator interface {
    Translate(key, locale string) string
}
```

`Translate` returns a localized string by key and locale. If the translation is not found, it returns the key unchanged.

**MapTranslator** — implementation based on nested maps:

```go
func NewMapTranslator(m map[string]map[string]string) *MapTranslator
```

Outer key — locale, inner key — translation key:

```go
tr := schema.NewMapTranslator(map[string]map[string]string{
    "uk": {
        "user.name":  "Ім'я",
        "user.email": "Електронна пошта",
    },
    "en": {
        "user.name":  "Name",
        "user.email": "Email",
    },
})
```

You can create your own `Translator` implementation, for example with `.po` file support, a database, or an external service.

---

## `parser` Package

### Generation from Go Structs

#### GenerateJSONSchema

```go
func GenerateJSONSchema(v any) (*schema.JSONSchema, error)
```

Generates a JSON Schema (Draft 7) from a Go value. Accepts a struct or a pointer to a struct.

#### GenerateJSONSchemaWithOptions

```go
func GenerateJSONSchemaWithOptions(v any, opts schema.Options) (*schema.JSONSchema, error)
```

Same as above, with the ability to specify options (Draft, Translator, etc.).

#### GenerateUISchema

```go
func GenerateUISchema(v any) (*schema.UISchemaElement, error)
```

Generates a UI Schema (JSON Forms) from a Go value.

#### GenerateUISchemaWithOptions

```go
func GenerateUISchemaWithOptions(v any, opts schema.Options) (*schema.UISchemaElement, error)
```

Same as above, with the ability to specify options (i18n, renderers, permissions).

**Generation logic:**

1. Iterates over all exported struct fields.
2. Fields tagged with `json:"-"` are skipped.
3. If `OmitEmpty: true` — fields with `json:",omitempty"` and a zero value are skipped (recursively for nested structs).
4. Nested structs (except `time.Time`) are processed recursively as `"object"` / `Group`.
5. `time.Time` is mapped as `"string"` with `format: "date-time"`.
6. Pointers (`*T`) are unwrapped to the base type.
7. For UI Schema:
   - `form:"hidden"` or `AccessHidden` → field is excluded.
   - `AccessReadOnly` → `options.readonly = true`.
   - Nested structs → `Group` with the field name as label. If the struct field has `form:"category=..."`, the Group is placed into the corresponding category.
   - Fields of type `[]struct` or `[]*struct` → `Control` with `options.detail` containing a `VerticalLayout` with Controls for the array item's fields. Scopes inside detail are relative to the item: `#/properties/fieldName`. Primitive slices (`[]string`, `[]int`, etc.) remain plain Controls without `options.detail`.
   - Categories → automatic wrapping into `Categorization` → `Category`.
   - Rules are applied by priority: `visibleIf` → `hideIf` → `enableIf` → `disableIf`.
   - Group rules: `visibleIf`/`hideIf`/`enableIf`/`disableIf` tags on a struct field apply to the corresponding `Group`.
   - Category rules: `visibleIf`/`hideIf`/`enableIf`/`disableIf` directives in the `form` tag apply to the `Category`.
   - Category i18n: `i18n=key` directive in the `form` tag sets the `i18n` field and translates the label via `Translator`.
   - `layout=horizontal`: adjacent fields with the same `layout=horizontal` are grouped into a `HorizontalLayout`.
   - `layout=horizontal:groupName`: non-adjacent fields with the same named group are combined into a single `HorizontalLayout` at the first-occurrence position. Single-element named groups remain as plain Controls.
8. Validation constraints (`minLength`, `maxLength`, `minimum`, `maximum`, `pattern`, `description`) from struct tags are written to JSON Schema.

---

### Generation from JSON Objects

#### GenerateFromJSON

```go
func GenerateFromJSON(data []byte) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Generates both schemas from raw JSON. Input data must be a JSON object. All fields are considered optional.

#### GenerateFromJSONWithOptions

```go
func GenerateFromJSONWithOptions(data []byte, opts schema.Options) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Same as above, with options.

**Sentinel errors:**

| Variable | Description |
|----------|-------------|
| `ErrInvalidJSON` | Input data is not valid JSON |
| `ErrNotAnObject` | Top-level JSON is not an object |

**JSON → JSON Schema type auto-detection:**

| JSON value | JSON Schema `type` |
|------------|---------------------|
| `null` | `"null"` |
| `true` / `false` | `"boolean"` |
| `42` (integer) | `"integer"` |
| `3.14` (float) | `"number"` |
| `"text"` | `"string"` |
| `[...]` | `"array"` (items type from first element) |
| `{...}` | `"object"` (properties recursively) |

---

### Generation from OpenAPI 3.x

#### GenerateFromOpenAPI

```go
func GenerateFromOpenAPI(data []byte, schemaName string) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Parses an OpenAPI 3.x JSON document and generates both schemas for the specified component. `schemaName` must match a key in `components.schemas`.

**Sentinel errors:**

| Variable | Description |
|----------|-------------|
| `ErrInvalidOpenAPI` | Document is not valid OpenAPI |
| `ErrSchemaNotFound` | Specified schema not found in `components.schemas` |

**Features:**

- Supports `$ref` references within the document (format: `#/components/schemas/Name`).
- Converts `type`, `properties`, `items`, `required`, `format`, `default`, `enum`, `description`, `title`.
- Nested objects with `properties` become `Group` in UI Schema.

---

## `api` Package

### Registry

A thread-safe registry of Go types for schema generation by name.

```go
type Registry struct { /* unexported fields */ }

func NewRegistry() *Registry
func (r *Registry) Register(name string, v any)
func (r *Registry) Lookup(name string) (any, error)
func (r *Registry) Names() []string
```

| Method | Description |
|--------|-------------|
| `NewRegistry()` | Creates an empty registry |
| `Register(name, v)` | Registers a struct instance under a name. Overwrites on duplicate. |
| `Lookup(name)` | Returns the registered instance or an error |
| `Names()` | Returns a list of all registered names |

**Example:**

```go
reg := api.NewRegistry()
reg.Register("User", User{})
reg.Register("Order", Order{})

names := reg.Names() // ["User", "Order"]
v, err := reg.Lookup("User") // User{}, nil
```

---

### Handler

HTTP handler for schema generation.

```go
type Handler struct { /* unexported fields */ }

func NewHandler(registry *Registry) *Handler
func (h *Handler) GenerateHandler(w http.ResponseWriter, r *http.Request)
```

**`MaxBodySize`** — request body size limit: `2 MB` (2 << 20).

---

### HTTP Endpoint

**`POST /schema/generate`**

Accepts a JSON body with one of two fields:

#### Option 1: by registered type

```json
{"type": "User"}
```

Looks up `"User"` in the registry and generates schemas from the Go struct.

#### Option 2: from raw JSON

```json
{"data": {"name": "John", "age": 30}}
```

Generates schemas from the provided JSON object.

#### Response (200 OK)

```json
{
  "schema": {
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": { "..." },
    "required": ["..."]
  },
  "uischema": {
    "type": "VerticalLayout",
    "elements": ["..."]
  }
}
```

#### Errors

| HTTP Code | Reason |
|-----------|--------|
| 405 | Not a POST method |
| 400 | Invalid JSON, missing `type` and `data` |
| 404 | Type not found in registry |
| 500 | Internal generation error |

Error format:

```json
{"error": "type 'Unknown' not found"}
```

---

## Struct Tags

### Base Tags

| Tag | Value | Effect on JSON Schema | Example |
|-----|-------|----------------------|---------|
| `json:"name"` | Field name | Key in `properties` | `json:"email"` |
| `json:"-"` | Skip | Field is excluded entirely | `json:"-"` |
| `json:",omitempty"` | Skip if empty | With `OmitEmpty: true` field is excluded at zero value | `json:"notes,omitempty"` |
| `required:"true"` | Required | Added to `required[]` | `required:"true"` |
| `default:"value"` | Default | Sets `default` (auto type coercion) | `default:"true"`, `default:"42"` |
| `enum:"a,b,c"` | Allowed values | Sets `enum[]` | `enum:"admin,user,moderator"` |
| `format:"fmt"` | Format | Sets `format` | `format:"email"`, `format:"date-time"`, `format:"uri"` |
| `description:"text"` | Description | Sets `description` | `description:"User name"` |
| `minLength:"n"` | Min length | Sets `minLength` | `minLength:"2"` |
| `maxLength:"n"` | Max length | Sets `maxLength` | `maxLength:"100"` |
| `minimum:"n"` | Min value | Sets `minimum` | `minimum:"0"`, `minimum:"1.5"` |
| `maximum:"n"` | Max value | Sets `maximum` | `maximum:"999"`, `maximum:"99.9"` |
| `pattern:"regex"` | Pattern | Sets `pattern` | `pattern:"^[A-Z]"` |

**`default` type coercion:**

| Go field type | Coercion |
|---------------|----------|
| `bool` | `"true"` → `true`, `"false"` → `false` |
| `int`, `int8`–`int64` | `"42"` → `42` |
| `uint`, `uint8`–`uint64` | `"10"` → `10` |
| `float32`, `float64` | `"3.14"` → `3.14` |
| `string` | No change |

### `form` Tag — UI Options

Directives are separated by `;`:

| Directive | Description | Example |
|-----------|-------------|---------|
| `label=Text` | Control label | `form:"label=Full Name"` |
| `hidden` | Hide from UI Schema | `form:"hidden"` |
| `readonly` | Read only | `form:"readonly"` |
| `multiline` | Multi-line field | `form:"multiline"` |
| `category=Name` | Category (tab) | `form:"category=Personal"` |
| `layout=horizontal` | Layout override (adjacent fields) | `form:"layout=horizontal"` |
| `layout=horizontal:groupName` | Named layout group (non-adjacent fields) | `form:"layout=horizontal:contact"` |
| `visibleIf=field:value` | SHOW rule for Category | `form:"category=Address;visibleIf=provideAddress:true"` |
| `hideIf=field:value` | HIDE rule for Category | `form:"category=Secret;hideIf=role:admin"` |
| `enableIf=field:value` | ENABLE rule for Category | `form:"category=Settings;enableIf=active:true"` |
| `disableIf=field:value` | DISABLE rule for Category | `form:"category=Edit;disableIf=locked:true"` |
| `i18n=key` | i18n key for Category | `form:"category=Personal;i18n=category.personal"` |

**Combinations:**

```go
type Profile struct {
    Name  string `json:"name" form:"label=Name;category=General"`
    Bio   string `json:"bio" form:"multiline;readonly;category=Additional"`
    Token string `json:"token" form:"hidden"`
}
```

### Conditional Rule Tags

Value format: `"field=value"`.

| Tag | Effect | Example | Result |
|-----|--------|---------|--------|
| `visibleIf:"field=val"` | `SHOW` | `visibleIf:"is_active=true"` | Show when `is_active == true` |
| `hideIf:"field=val"` | `HIDE` | `hideIf:"role=admin"` | Hide when `role == "admin"` |
| `enableIf:"field=val"` | `ENABLE` | `enableIf:"agreed=true"` | Enable when `agreed == true` |
| `disableIf:"field=val"` | `DISABLE` | `disableIf:"locked=true"` | Disable when `locked == true` |

> **Note:** These tags work on both regular fields (Control) and nested struct fields (Group). For Category rules, use the `visibleIf`, `hideIf`, `enableIf`, `disableIf` directives inside the `form` tag (see [form Tag](#form-tag--ui-options)).

**Rule priority** (only the first match is applied):

`visibleIf` → `hideIf` → `enableIf` → `disableIf`

**Value auto-coercion:**

- `"true"` / `"false"` → `bool`
- `"42"` → `int`
- `"3.14"` → `float64`
- Everything else → `string`

**Result in JSON:**

```json
{
  "type": "Control",
  "scope": "#/properties/details",
  "rule": {
    "effect": "SHOW",
    "condition": {
      "scope": "#/properties/is_active",
      "schema": { "const": true }
    }
  }
}
```

### i18n and renderer

| Tag | Description | Example |
|-----|-------------|---------|
| `i18n:"key"` | Translation key for the label | `i18n:"user.name"` |
| `renderer:"name"` | Custom renderer | `renderer:"color-picker"` |

If a `Translator` is set in `Options` and a translation exists for the `i18n` key — the control label is replaced with the translation. If no translation is found — `form:"label=..."` or the Go field name is used.

The tag renderer takes priority over `Options.Renderers`.

---

## Type Mapping

| Go type | JSON Schema `type` | Additional |
|---------|---------------------|------------|
| `string` | `"string"` | — |
| `bool` | `"boolean"` | — |
| `int`, `int8`, `int16`, `int32`, `int64` | `"integer"` | — |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `"integer"` | — |
| `float32`, `float64` | `"number"` | — |
| `time.Time` | `"string"` | `format: "date-time"` |
| `[]T`, `[N]T` | `"array"` | `items` — schema of `T` |
| `map[string]T` | `"object"` | `additionalProperties` — schema of `T` |
| `map[K]T` (K ≠ string) | `"object"` | No `additionalProperties` |
| nested `struct` | `"object"` | `properties` recursively |
| `*T` (pointer) | unwrapped to `T` | — |
| other types | `"string"` | Fallback |

---

## Usage Examples

### Basic Generation

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/holdemlab/ui-json-schema/parser"
)

type User struct {
    ID       int    `json:"id" form:"hidden"`
    Name     string `json:"name" required:"true" form:"label=Full name"`
    Email    string `json:"email" required:"true" format:"email"`
    IsActive bool   `json:"is_active" default:"true"`
    Role     string `json:"role" enum:"admin,user,moderator"`
    Bio      string `json:"bio" form:"multiline"`
}

func main() {
    // JSON Schema
    jsonSchema, err := parser.GenerateJSONSchema(User{})
    if err != nil {
        panic(err)
    }

    // UI Schema
    uiSchema, err := parser.GenerateUISchema(User{})
    if err != nil {
        panic(err)
    }

    s, _ := json.MarshalIndent(jsonSchema, "", "  ")
    fmt.Println("JSON Schema:", string(s))

    u, _ := json.MarshalIndent(uiSchema, "", "  ")
    fmt.Println("UI Schema:", string(u))
}
```

**JSON Schema result:**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "id":        { "type": "integer" },
    "name":      { "type": "string" },
    "email":     { "type": "string", "format": "email" },
    "is_active": { "type": "boolean", "default": true },
    "role":      { "type": "string", "enum": ["admin", "user", "moderator"] },
    "bio":       { "type": "string" }
  },
  "required": ["name", "email"]
}
```

**UI Schema result:**

```json
{
  "type": "VerticalLayout",
  "elements": [
    { "type": "Control", "scope": "#/properties/name", "label": "Full name" },
    { "type": "Control", "scope": "#/properties/email" },
    { "type": "Control", "scope": "#/properties/is_active" },
    { "type": "Control", "scope": "#/properties/role" },
    { "type": "Control", "scope": "#/properties/bio", "options": { "multi": true } }
  ]
}
```

> Note: the `id` field is absent from UI Schema due to `form:"hidden"`.

---

### i18n — Localization

```go
package main

import (
    "github.com/holdemlab/ui-json-schema/parser"
    "github.com/holdemlab/ui-json-schema/schema"
)

type User struct {
    Name  string `json:"name" i18n:"user.name" form:"label=Name"`
    Email string `json:"email" i18n:"user.email"`
}

func main() {
    tr := schema.NewMapTranslator(map[string]map[string]string{
        "uk": {
            "user.name":  "Ім'я",
            "user.email": "Електронна пошта",
        },
        "en": {
            "user.name":  "Name",
            "user.email": "Email Address",
        },
    })

    opts := schema.Options{
        Translator: tr,
        Locale:     "uk",
    }

    ui, _ := parser.GenerateUISchemaWithOptions(User{}, opts)
    // Control "name"  → label: "Ім'я"
    // Control "email" → label: "Електронна пошта"
}
```

**Custom Translator implementation:**

```go
type DBTranslator struct {
    db *sql.DB
}

func (t *DBTranslator) Translate(key, locale string) string {
    var val string
    err := t.db.QueryRow(
        "SELECT value FROM translations WHERE key = $1 AND locale = $2",
        key, locale,
    ).Scan(&val)
    if err != nil {
        return key // fallback to key
    }
    return val
}
```

---

### Custom Renderers

Renderers let you tell JSON Forms which UI component to use for a field.

**Via struct tag (takes priority):**

```go
type Config struct {
    Color  string `json:"color" renderer:"color-picker"`
    Rating int    `json:"rating" renderer:"star-rating"`
}

ui, _ := parser.GenerateUISchema(Config{})
```

Result:

```json
{
  "type": "Control",
  "scope": "#/properties/color",
  "options": { "renderer": "color-picker" }
}
```

**Via Options (by scope):**

```go
opts := schema.Options{
    Renderers: map[string]string{
        "#/properties/rating": "star-rating",
        "#/properties/avatar": "image-upload",
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Config{}, opts)
```

> The `renderer` tag takes priority over `Options.Renderers`.

---

### Roles and Permissions

```go
type Article struct {
    Title   string `json:"title"`
    Content string `json:"content"`
    Status  string `json:"status"`
    Author  string `json:"author"`
}

opts := schema.Options{
    Role: "editor",
    RolePermissions: map[string]schema.FieldPermissions{
        "viewer": {
            "title":   schema.AccessReadOnly,
            "content": schema.AccessReadOnly,
            "status":  schema.AccessHidden,
            "author":  schema.AccessReadOnly,
        },
        "editor": {
            "title":   schema.AccessFull,
            "content": schema.AccessFull,
            "status":  schema.AccessReadOnly,  // can view, can't edit
            "author":  schema.AccessHidden,     // completely hidden
        },
        "admin": {
            // all fields with full access (AccessFull by default)
        },
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Article{}, opts)
// title   → regular control
// content → regular control
// status  → readonly control
// author  → absent from UI Schema
```

---

### Categorization (Tabs)

When at least one field has a `form:"category=..."` tag, the root layout automatically becomes `Categorization`, and fields are grouped into corresponding `Category` elements.

```go
type RegistrationForm struct {
    FirstName string `json:"first_name" form:"label=First Name;category=Personal Data"`
    LastName  string `json:"last_name" form:"label=Last Name;category=Personal Data"`
    Email     string `json:"email" form:"category=Contacts" format:"email"`
    Phone     string `json:"phone" form:"category=Contacts"`
    Company   string `json:"company" form:"category=Work"`
    Position  string `json:"position" form:"category=Work"`
}

ui, _ := parser.GenerateUISchema(RegistrationForm{})
```

Result:

```json
{
  "type": "Categorization",
  "elements": [
    {
      "type": "Category",
      "label": "Personal Data",
      "elements": [
        { "type": "Control", "scope": "#/properties/first_name", "label": "First Name" },
        { "type": "Control", "scope": "#/properties/last_name", "label": "Last Name" }
      ]
    },
    {
      "type": "Category",
      "label": "Contacts",
      "elements": [
        { "type": "Control", "scope": "#/properties/email" },
        { "type": "Control", "scope": "#/properties/phone" }
      ]
    },
    {
      "type": "Category",
      "label": "Work",
      "elements": [
        { "type": "Control", "scope": "#/properties/company" },
        { "type": "Control", "scope": "#/properties/position" }
      ]
    }
  ]
}
```

---

### Nested Structs in Categories

When a struct field has `form:"category=..."`, the corresponding `Group` is placed into the specified category instead of "Other":

```go
type DrawSetup struct {
    Numbers int `json:"numbers"`
    Bonus   int `json:"bonus"`
}

type GameConfig struct {
    GameName  string    `json:"game_name" form:"label=Game name;category=General"`
    DrawSetup DrawSetup `json:"draw_setup" form:"label=Draw setup;category=General"`
    Logic     string    `json:"logic" form:"label=Logic;category=Logic"`
}

ui, _ := parser.GenerateUISchema(GameConfig{})
```

Result — `DrawSetup` is rendered as a `Group` inside the "General" category:

```json
{
  "type": "Categorization",
  "elements": [
    {
      "type": "Category",
      "label": "General",
      "elements": [
        { "type": "Control", "scope": "#/properties/game_name", "label": "Game name" },
        {
          "type": "Group",
          "label": "Draw setup",
          "elements": [
            { "type": "Control", "scope": "#/properties/draw_setup/properties/numbers" },
            { "type": "Control", "scope": "#/properties/draw_setup/properties/bonus" }
          ]
        }
      ]
    },
    {
      "type": "Category",
      "label": "Logic",
      "elements": [
        { "type": "Control", "scope": "#/properties/logic", "label": "Logic" }
      ]
    }
  ]
}
```

> **Note:** Without `form:"category=..."` on the struct field, the `Group` falls into the "Other" category. Rules (`visibleIf`, `hideIf`, etc.) and `i18n` in the `form` tag also work for nested structs in categories.

---

### Conditional Visibility

```go
type Survey struct {
    HasPet    bool   `json:"has_pet"`
    PetName   string `json:"pet_name" visibleIf:"has_pet=true" form:"label=Pet Name"`
    PetAge    int    `json:"pet_age" visibleIf:"has_pet=true"`
    Country   string `json:"country"`
    State     string `json:"state" enableIf:"country=US" form:"label=State"`
    IsMinor   bool   `json:"is_minor"`
    ParentName string `json:"parent_name" visibleIf:"is_minor=true"`
    Reason    string `json:"reason" hideIf:"has_pet=false"`
}

ui, _ := parser.GenerateUISchema(Survey{})
```

The `pet_name` control in UI Schema:

```json
{
  "type": "Control",
  "scope": "#/properties/pet_name",
  "label": "Pet Name",
  "rule": {
    "effect": "SHOW",
    "condition": {
      "scope": "#/properties/has_pet",
      "schema": { "const": true }
    }
  }
}
```

---

### Validation Constraints

Struct tags `minLength`, `maxLength`, `minimum`, `maximum`, `pattern`, `description` are written to JSON Schema:

```go
type Registration struct {
    Username string  `json:"username" minLength:"3" maxLength:"20" pattern:"^[a-zA-Z0-9_]+$" description:"User login"`
    Email    string  `json:"email" format:"email" description:"Email address"`
    Age      int     `json:"age" minimum:"18" maximum:"120"`
    Score    float64 `json:"score" minimum:"0" maximum:"100"`
}

js, _ := parser.GenerateJSONSchema(Registration{})
```

Result for `username`:

```json
{
  "type": "string",
  "description": "User login",
  "minLength": 3,
  "maxLength": 20,
  "pattern": "^[a-zA-Z0-9_]+$"
}
```

Result for `age`:

```json
{
  "type": "integer",
  "minimum": 18,
  "maximum": 120
}
```

---

### HorizontalLayout

Adjacent fields with the same `form:"layout=horizontal"` are automatically grouped into a `HorizontalLayout`:

```go
type Address struct {
    City    string `json:"city" form:"layout=horizontal"`
    ZipCode string `json:"zip_code" form:"layout=horizontal"`
    Country string `json:"country"`
}

ui, _ := parser.GenerateUISchema(Address{})
```

Resulting UI Schema:

```json
{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "HorizontalLayout",
      "elements": [
        { "type": "Control", "scope": "#/properties/city" },
        { "type": "Control", "scope": "#/properties/zip_code" }
      ]
    },
    { "type": "Control", "scope": "#/properties/country" }
  ]
}
```

---

### Named Layout Groups

The syntax `form:"layout=horizontal:groupName"` allows combining **non-adjacent** fields with the same group name into a single `HorizontalLayout`. Fields are grouped at the first-occurrence position of the group.

If a named group contains only a single element, it remains a plain `Control` without being wrapped in a `HorizontalLayout`.

The `layout` and `layoutGroup` options are removed from elements after grouping.

```go
type ContactForm struct {
    FirstName  string `json:"first_name" form:"layout=horizontal:name"`
    Email      string `json:"email" form:"layout=horizontal:contact"`
    LastName   string `json:"last_name" form:"layout=horizontal:name"`
    Phone      string `json:"phone" form:"layout=horizontal:contact"`
    MiddleName string `json:"middle_name"`
}

ui, _ := parser.GenerateUISchema(ContactForm{})
```

Resulting UI Schema — `first_name` and `last_name` (group `name`) and `email` and `phone` (group `contact`) are combined into `HorizontalLayout` at the first-occurrence position of each group:

```json
{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "HorizontalLayout",
      "elements": [
        { "type": "Control", "scope": "#/properties/first_name" },
        { "type": "Control", "scope": "#/properties/last_name" }
      ]
    },
    {
      "type": "HorizontalLayout",
      "elements": [
        { "type": "Control", "scope": "#/properties/email" },
        { "type": "Control", "scope": "#/properties/phone" }
      ]
    },
    { "type": "Control", "scope": "#/properties/middle_name" }
  ]
}
```

> **Note:** Unnamed `layout=horizontal` (without `:groupName`) still works as before — grouping only adjacent fields.

---

### Rules on Category

The `visibleIf`, `hideIf`, `enableIf`, `disableIf` directives in the `form` tag allow adding a `Rule` to a `Category`:

```go
type Profile struct {
    Name            string `json:"name" form:"category=Personal"`
    ProvideAddress  bool   `json:"provideAddress" form:"category=Personal"`
    Street          string `json:"street" form:"category=Address;visibleIf=provideAddress:true"`
    City            string `json:"city" form:"category=Address;visibleIf=provideAddress:true"`
}

ui, _ := parser.GenerateUISchema(Profile{})
```

The "Address" category will have a rule:

```json
{
  "type": "Category",
  "label": "Address",
  "elements": [...],
  "rule": {
    "effect": "SHOW",
    "condition": {
      "scope": "#/properties/provideAddress",
      "schema": { "const": true }
    }
  }
}
```

---

### i18n on Category

The `i18n=key` directive in the `form` tag sets the `i18n` field on a `Category` and translates the label via `Translator`:

```go
type Settings struct {
    Name  string `json:"name" form:"category=Personal;i18n=category.personal"`
    Theme string `json:"theme" form:"category=Appearance;i18n=category.appearance"`
}

opts := schema.Options{
    Translator: func(key string) string {
        translations := map[string]string{
            "category.personal":   "Особисте",
            "category.appearance": "Зовнішній вигляд",
        }
        return translations[key]
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Settings{}, opts)
```

Result:

```json
{
  "type": "Category",
  "label": "Особисте",
  "i18n": "category.personal",
  "elements": [...]
}
```

---

### Rules on Group (Nested Structs)

The `visibleIf`, `hideIf`, `enableIf`, `disableIf` tags on a struct field add a rule to the corresponding `Group`:

```go
type Order struct {
    Total          float64 `json:"total"`
    ProvideAddress bool    `json:"provideAddress"`
    Address        struct {
        Street string `json:"street"`
        City   string `json:"city"`
    } `json:"address" visibleIf:"provideAddress=true"`
}

ui, _ := parser.GenerateUISchema(Order{})
```

The "address" group will have a rule:

```json
{
  "type": "Group",
  "label": "Address",
  "elements": [
    { "type": "Control", "scope": "#/properties/address/properties/street" },
    { "type": "Control", "scope": "#/properties/address/properties/city" }
  ],
  "rule": {
    "effect": "SHOW",
    "condition": {
      "scope": "#/properties/provideAddress",
      "schema": { "const": true }
    }
  }
}
```

---

### Array Detail (Slice of Structs)

When a field has type `[]SomeStruct` or `[]*SomeStruct`, the UI Schema generator produces a `Control` with `options.detail` containing a `VerticalLayout` with Controls for the item struct's fields. This follows the JSON Forms convention for array controls.

- Scopes inside `options.detail` are relative to the array item: `#/properties/fieldName`
- All existing features work inside detail: labels, readonly, multiline, rules, horizontal layout, nested structs (Groups)
- Primitive slices (`[]string`, `[]int`, etc.) are unchanged — they remain plain Controls without `options.detail`
- Empty structs don't produce a detail
- Horizontal grouping (`groupHorizontalElements`) is also supported inside array items

```go
type WinningSet struct {
    Numbers int    `json:"numbers"`
    Bonus   int    `json:"bonus"`
    Label   string `json:"label" form:"label=Set Label"`
}

type GameConfig struct {
    GameName    string       `json:"game_name" form:"label=Game name"`
    WinningSets []WinningSet `json:"winning_sets" form:"label=Winning Sets"`
}
```

**UI Schema:**

```json
{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "Control",
      "label": "Game name",
      "scope": "#/properties/game_name"
    },
    {
      "type": "Control",
      "label": "Winning Sets",
      "scope": "#/properties/winning_sets",
      "options": {
        "detail": {
          "type": "VerticalLayout",
          "elements": [
            { "type": "Control", "scope": "#/properties/numbers" },
            { "type": "Control", "scope": "#/properties/bonus" },
            { "type": "Control", "label": "Set Label", "scope": "#/properties/label" }
          ]
        }
      }
    }
  ]
}
```

---

### JSON Schema Draft 2019-09

```go
opts := schema.Options{Draft: "2019-09"}

jsonSchema, _ := parser.GenerateJSONSchemaWithOptions(User{}, opts)
// $schema → "https://json-schema.org/draft/2019-09/schema"
```

---

### OmitEmpty — Empty Field Filtering

When `OmitEmpty: true`, fields tagged with `json:",omitempty"` are excluded from the schema if their value is the zero value for their type. This works for both schemas — JSON Schema and UI Schema, and recursively for nested structs.

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/holdemlab/ui-json-schema/parser"
    "github.com/holdemlab/ui-json-schema/schema"
)

type Article struct {
    Title   string   `json:"title" required:"true"`
    Content string   `json:"content"`
    Notes   string   `json:"notes,omitempty"`
    Tags    []string `json:"tags,omitempty"`
    Views   int      `json:"views,omitempty"`
}

func main() {
    opts := schema.Options{
        Draft:     "draft-07",
        OmitEmpty: true,
    }

    // Empty struct — omitempty fields are excluded
    empty := Article{Title: "Hello"}
    s1, _ := parser.GenerateJSONSchemaWithOptions(empty, opts)
    b1, _ := json.MarshalIndent(s1, "", "  ")
    fmt.Println("Empty:", string(b1))
    // properties: title, content (without notes, tags, views)

    // Populated struct — omitempty fields are included
    full := Article{
        Title:   "Hello",
        Content: "World",
        Notes:   "draft",
        Tags:    []string{"go"},
        Views:   42,
    }
    s2, _ := parser.GenerateJSONSchemaWithOptions(full, opts)
    b2, _ := json.MarshalIndent(s2, "", "  ")
    fmt.Println("Full:", string(b2))
    // properties: title, content, notes, tags, views (all present)
}
```

**Result for empty struct:**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "title":   { "type": "string" },
    "content": { "type": "string" }
  },
  "required": ["title"]
}
```

> Fields `notes`, `tags`, and `views` are absent because they have `omitempty` and zero values.

**Zero values by type:**

| Type | Zero value |
|------|------------|
| `string` | `""` |
| `int`, `float`, etc. | `0` |
| `bool` | `false` |
| `slice`, `map` | `nil` |
| `*T` (pointer) | `nil` |
| `struct` | all fields are zero |

> **Note:** without `OmitEmpty: true` (default) all fields with `omitempty` are always included in the schema.

---

### OpenAPI 3.x → JSON Forms

```go
openAPIDoc := []byte(`{
    "openapi": "3.0.0",
    "components": {
        "schemas": {
            "Pet": {
                "type": "object",
                "required": ["name"],
                "properties": {
                    "name": {
                        "type": "string",
                        "description": "Pet name"
                    },
                    "age": {
                        "type": "integer"
                    },
                    "vaccinated": {
                        "type": "boolean",
                        "default": false
                    },
                    "owner": {
                        "$ref": "#/components/schemas/Owner"
                    }
                }
            },
            "Owner": {
                "type": "object",
                "properties": {
                    "name":  { "type": "string" },
                    "email": { "type": "string", "format": "email" }
                }
            }
        }
    }
}`)

jsonSchema, uiSchema, err := parser.GenerateFromOpenAPI(openAPIDoc, "Pet")
// $ref "#/components/schemas/Owner" is resolved automatically
```

---

### Generation from JSON Data

```go
data := []byte(`{
    "name": "John",
    "age": 30,
    "is_active": true,
    "scores": [95, 87, 92],
    "address": {
        "city": "New York",
        "street": "Broadway",
        "zip": "10001"
    }
}`)

jsonSchema, uiSchema, err := parser.GenerateFromJSON(data)
if err != nil {
    // err == parser.ErrInvalidJSON or parser.ErrNotAnObject
    panic(err)
}
```

Result — JSON Schema with automatically detected types:
- `"name"` → `string`
- `"age"` → `integer` (whole number)
- `"is_active"` → `boolean`
- `"scores"` → `array` with `items: {type: "integer"}`
- `"address"` → `object` with `properties`

---

### HTTP Server with Type Registry

```go
package main

import (
    "log"
    "net/http"

    handler "github.com/holdemlab/ui-json-schema/api"
)

type User struct {
    Name     string `json:"name" required:"true" form:"label=Name"`
    Email    string `json:"email" required:"true" format:"email"`
    IsActive bool   `json:"is_active" default:"true"`
}

type Product struct {
    Title    string  `json:"title" required:"true"`
    Price    float64 `json:"price" required:"true"`
    Category string  `json:"category" enum:"electronics,clothing,food"`
}

func main() {
    registry := handler.NewRegistry()
    registry.Register("User", User{})
    registry.Register("Product", Product{})

    h := handler.NewHandler(registry)

    mux := http.NewServeMux()
    mux.HandleFunc("/schema/generate", h.GenerateHandler)

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

**Requests:**

```bash
# Generate from type
curl -s -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"type": "User"}' | jq .

# Generate from JSON
curl -s -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"data": {"title": "Laptop", "price": 999.99}}' | jq .
```

---

### Combined Example

A complete example with all features:

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/holdemlab/ui-json-schema/parser"
    "github.com/holdemlab/ui-json-schema/schema"
)

type Employee struct {
    // General
    ID        int    `json:"id" form:"hidden"`
    FirstName string `json:"first_name" required:"true" form:"label=First Name;category=General" i18n:"emp.first_name"`
    LastName  string `json:"last_name" required:"true" form:"label=Last Name;category=General" i18n:"emp.last_name"`
    Email     string `json:"email" required:"true" format:"email" form:"category=General"`

    // Work
    Department string `json:"department" enum:"engineering,marketing,sales,hr" form:"category=Work"`
    Position   string `json:"position" form:"category=Work"`
    Salary     int    `json:"salary" form:"category=Work;readonly"`

    // Additional
    Bio       string `json:"bio" form:"multiline;category=Additional"`
    IsRemote  bool   `json:"is_remote" default:"false" form:"category=Additional"`
    Office    string `json:"office" hideIf:"is_remote=true" form:"category=Additional"`
    Equipment string `json:"equipment" visibleIf:"is_remote=true" form:"category=Additional" renderer:"equipment-selector"`
}

func main() {
    tr := schema.NewMapTranslator(map[string]map[string]string{
        "uk": {
            "emp.first_name": "Ім'я",
            "emp.last_name":  "Прізвище",
        },
    })

    opts := schema.Options{
        Translator: tr,
        Locale:     "uk",
        Draft:      "draft-07",
        Role:       "manager",
        RolePermissions: map[string]schema.FieldPermissions{
            "manager": {
                "salary": schema.AccessReadOnly,
            },
            "hr": {
                // full access to all fields
            },
            "employee": {
                "salary":     schema.AccessHidden,
                "department": schema.AccessReadOnly,
            },
        },
    }

    jsonSchema, _ := parser.GenerateJSONSchemaWithOptions(Employee{}, opts)
    uiSchema, _ := parser.GenerateUISchemaWithOptions(Employee{}, opts)

    s, _ := json.MarshalIndent(jsonSchema, "", "  ")
    u, _ := json.MarshalIndent(uiSchema, "", "  ")

    fmt.Println("=== JSON Schema ===")
    fmt.Println(string(s))
    fmt.Println("\n=== UI Schema ===")
    fmt.Println(string(u))
}
```

---

## HTTP API — Reference

### Starting the Built-in Server

```bash
# Default :8080
make run

# Or with a custom port
ADDR=":3000" go run cmd/server/main.go
```

### Endpoint

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/schema/generate` | Generate JSON Schema + UI Schema |

### Request Format

**Content-Type:** `application/json`  
**Maximum body size:** 2 MB

**Fields (one of two):**

| Field | Type | Description |
|-------|------|-------------|
| `type` | `string` | Name of a registered Go type |
| `data` | `object` | Raw JSON object for analysis |

### Response Format

**Success (200):**

```json
{
  "schema":   { /* JSONSchema */ },
  "uischema": { /* UISchemaElement */ }
}
```

**Error (4xx/5xx):**

```json
{
  "error": "error message"
}
```

### Error Codes

| Code | Reason |
|------|--------|
| `405 Method Not Allowed` | Non-POST method used |
| `400 Bad Request` | Invalid JSON or missing required fields |
| `404 Not Found` | Type not found in registry |
| `500 Internal Server Error` | Schema generation error |

---

## Performance

Benchmarks on Intel i7-14700HX:

| Operation | Time | Allocations |
|-----------|------|-------------|
| JSON Schema from small struct (5 fields) | ~2.2 µs | 9 allocs |
| JSON Schema from medium struct (15 fields) | ~10.4 µs | 42 allocs |
| JSON Schema from large struct (40+ fields) | ~25.5 µs | 93 allocs |
| Generation from 1 MB JSON | ~3.9 ms | 3,536 allocs |
| Generation from 2 MB JSON | ~6.3 ms | 3,536 allocs |

All operations are well below the 100 ms target for JSON up to 2 MB.

---

## Development

### Requirements

- Go 1.24+
- golangci-lint v2
- GNU Make

### Make Commands

| Command | Description |
|---------|-------------|
| `make test` | Run tests |
| `make test-cover` | Tests with coverage (minimum 80%) |
| `make bench` | Benchmarks |
| `make lint` | Run linter |
| `make build` | Build server |
| `make run` | Start server |

### CI Pipeline

1. **Lint** — golangci-lint v2 with `.golangci.yml` configuration
2. **Test** — all tests + coverage check ≥ 80%
3. **Build** — `go build ./...`
4. **Tag** — automatic tag creation (master branch only) via Conventional Commits

### Conventional Commits

| Format | Version |
|--------|---------|
| `BREAKING CHANGE:` in body | Major (v1.0.0 → v2.0.0) |
| `feat: ...` | Minor (v0.1.0 → v0.2.0) |
| `fix: ...`, `docs: ...`, `chore: ...` | Patch (v0.1.0 → v0.1.1) |

---

## License

MIT
