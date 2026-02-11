# ROADMAP — JSON Schema & UI Schema Generator

Automatic generation of JSON Schema and UI Schema from Go structs and JSON objects for JSON Forms.

---

## Stage 0 — Project Initialization ✅

- [x] Go module initialization (`go mod init`)
- [x] Directory structure setup (`schema/`, `parser/`, `api/`, `cmd/server/`)
- [x] GolangCI-Lint v2 configuration (base set of linters)
- [x] CI setup (GitHub Actions: lint + test + build)
- [x] `Makefile` creation (lint, test, test-cover, build, fmt, clean)
- [x] `.gitignore` creation

**Result:** empty project with a working CI pipeline.

---

## Stage 1 — Core: JSON Schema Generation from Go Structs ✅

- [x] `schema/jsonschema.go` — basic JSON Schema types (Draft 7)
- [x] `parser/struct_parser.go` — Go struct analysis via `reflect`
- [x] Support for primitive types: `string`, `int`, `int32`, `int64`, `float32`, `float64`, `bool`
- [x] Support for `time.Time` → `{"type":"string","format":"date-time"}`
- [x] Support for nested `struct`
- [x] Support for `slices` (`[]T`) → `{"type":"array","items":{…}}`
- [x] Support for `map[string]T` → `{"type":"object","additionalProperties":{…}}`
- [x] Reading `json` tag for field name
- [x] Unit tests (coverage 91.3% ≥ 80%)

**Result:** `GenerateJSONSchema(v any) (JSONSchema, error)` — generates a valid JSON Schema from any Go struct.

---

## Stage 2 — Struct Tag Support ✅

- [x] `schema/tags.go` — custom tag parsing
- [x] Tag `required:"true"` → field added to `required`
- [x] Tag `default:"…"` → `"default": …`
- [x] Tag `enum:"a,b,c"` → `"enum": ["a","b","c"]`
- [x] Tag `format:"email"` / `format:"date"` / `format:"date-time"` → `"format": "…"`
- [x] Handling `json:"-"` (skip field)
- [x] Handling `json:",omitempty"`
- [x] Unit tests for each tag (coverage 91.8%)

**Result:** JSON Schema takes all declared tags into account.

---

## Stage 3 — UI Schema Generation ✅

- [x] `schema/uischema.go` — UI Schema types (JSON Forms)
- [x] Automatic creation of `VerticalLayout` with `Control` elements
- [x] `scope` → `#/properties/<field>`
- [x] Parsing `form:"…"` tag:
  - `label=Full name` → `"label": "Full name"`
  - `hidden` → element not added to UI Schema
  - `readonly` → `"options": {"readonly": true}`
  - `multiline` → `"options": {"multi": true}`
- [x] Recursive processing of nested structs (Group / nested layout)
- [x] Unit tests (coverage 92.4%)

**Result:** `GenerateUISchema(v any) (UISchema, error)` — generates a UI Schema compatible with JSON Forms.

---

## Stage 4 — Rules (Conditional Logic)

- [x] Parsing `visibleIf`, `hideIf`, `enableIf`, `disableIf` tags
- [x] Generating `rule` block in UI Schema:
  - `effect`: `SHOW` / `HIDE` / `ENABLE` / `DISABLE`
  - `condition`: `scope` + `schema.const`
- [x] Support for different value types in conditions (`bool`, `string`, `int`, `float`)
- [x] Rule priority: `visibleIf` → `hideIf` → `enableIf` → `disableIf`
- [x] Integration into `buildUIElements` via `applyRule()`
- [x] Unit tests (coverage 93.6%)

**Result:** UI Schema with conditional field display rules.

---

## Stage 5 — JSON Schema Generation from Arbitrary JSON

- [x] `parser/json_parser.go` — parsing `[]byte` → `map[string]any`
- [x] Value type inference (`string`, `number`, `integer`, `boolean`, `null`)
- [x] Distinguishing `integer` vs `number` (via `math.Trunc`)
- [x] Nested objects → nested `properties` + Group in UI Schema
- [x] Arrays → `items` (type inferred from first element)
- [x] Empty arrays → empty `items` schema
- [x] Object arrays → `items.properties` from first element
- [x] UI Schema generation from JSON object (VerticalLayout, Controls, Groups)
- [x] All fields `optional` (no `required`)
- [x] Validation: error for non-object JSON (array, string, number, null)
- [x] Unit tests (coverage 94.4%)

**Result:** `GenerateFromJSON(data []byte) (*JSONSchema, *UISchemaElement, error)` — generates both schemas from arbitrary JSON.

---

## Stage 6 — HTTP API

- [x] `api/registry.go` — Go type registry (`Registry`) with `Register`, `Lookup`, `Names`; thread-safe via `sync.RWMutex`
- [x] `api/handler.go` — HTTP handler `POST /schema/generate` (`GenerateHandler`)
- [x] Accepting `{"type":"Name"}` → generation from registered Go type
- [x] Accepting `{"data":{…}}` → generation from JSON object
- [x] Priority: `type` > `data` when both fields are present
- [x] Response format: `{"schema":{…},"uischema":{…}}`
- [x] Validation: errors for invalid JSON, empty body, missing type/data, unknown type, array instead of object
- [x] Correct HTTP status codes: 200 OK, 400 Bad Request, 404 Not Found, 405 Method Not Allowed
- [x] `cmd/server/main.go` — HTTP server with `http.ListenAndServe`, address configuration via `ADDR` env
- [x] Body limit 2 MB (`maxRequestBody`)
- [x] Integration tests for API (20+ tests: registry + handler) — API coverage 93.9%
- [x] Unit tests (total coverage 91.8%)

**Result:** working HTTP server that returns JSON Schema + UI Schema.

---

## Stage 7 — Performance & Quality ✅

- [x] Generation benchmarks (JSON up to 1–2 MB < 100 ms)
  - Small struct (5 fields): ~2.2 µs
  - Medium struct (15+ fields): ~10.4 µs
  - Large struct (40+ fields): ~25.5 µs
  - 1 MB JSON: ~3.9 ms ✅ (< 100 ms)
  - 2 MB JSON: ~6.3 ms ✅ (< 100 ms)
- [x] Profiling and optimization — not needed (all benchmarks well below 100 ms)
- [x] Test coverage check ≥ 80% — total coverage **91.8%** (parser 95.1%, schema 94.3%, API 93.9%)
- [x] Generated schema compatibility with JSON Forms — `parser/compatibility_test.go` (3 tests: StructSchema, UISchema, FromJSON)
- [x] Final linter pass — **0 issues**
- [x] README with usage examples — `README.md` (features, installation, quick start, tags table, type mapping, project structure, benchmarks)
- [x] Makefile `bench` target for running benchmarks

**Result:** production-ready library with documentation.

---

## Stage 8 — Extensions ✅

- [x] i18n for labels — `schema.Translator` interface, `MapTranslator` implementation, `i18n` struct tag, automatic label translation
- [x] Custom renderers mapping — `renderer` struct tag + `Options.Renderers` map (tag takes priority)
- [x] Permissions / readonly by role — `Options.Role` + `Options.RolePermissions`, access levels: ReadWrite / ReadOnly / Hidden
- [x] OpenAPI → JSON Forms — `parser.GenerateFromOpenAPI()`, support for `$ref`, nested objects, arrays, enums
- [x] JSON Schema Draft 2019-09 support — `Options.Draft`, `DraftURL()`, both parsers (struct + JSON)
- [x] Custom layouts (Categorization) — `form:"category=..."` tag, automatic grouping into Categorization/Category, "Other" fallback

**New files:**
- `schema/i18n.go` — Translator interface and MapTranslator
- `schema/options.go` — Options struct (Draft, Translator, Renderers, RolePermissions)
- `parser/openapi_parser.go` — OpenAPI 3.x → JSON Schema + UI Schema

**Test coverage:** 92.4% total (parser 93.9%, schema 95.4%, API 93.9%)
**Lint:** 0 issues

---

## Summary Table

| Stage | Name                           | Priority  | Dependency  |
|-------|--------------------------------|-----------|-------------|
| 0     | Project Initialization         | 🔴 High   | —           |
| 1     | JSON Schema from Go Structs    | 🔴 High   | Stage 0     |
| 2     | Struct Tag Support             | 🔴 High   | Stage 1     |
| 3     | UI Schema Generation           | 🔴 High   | Stage 1     |
| 4     | Rules (Conditional Logic)      | 🟡 Medium | Stage 3     |
| 5     | Generation from JSON           | 🟡 Medium | Stage 1     |
| 6     | HTTP API                       | 🟡 Medium | Stages 1-5  |
| 7     | Performance & Quality          | 🟡 Medium | Stages 1-6  |
| 8     | Extensions                     | 🟢 Low    | Stage 7     |
