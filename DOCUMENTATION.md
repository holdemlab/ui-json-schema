# ui-json-schema — Повна документація

> Бібліотека для автоматичної генерації **JSON Schema (Draft 7 / 2019-09)** та **UI Schema (JSON Forms)** з Go-структур, JSON-об'єктів та OpenAPI 3.x документів.

**Модуль:** `github.com/holdemlab/ui-json-schema`  
**Go:** 1.24+  
**Залежності:** відсутні (stdlib only)

---

## Зміст

1. [Встановлення](#встановлення)
2. [Архітектура](#архітектура)
3. [Пакет `schema` — типи та конфігурація](#пакет-schema)
   - [JSONSchema](#jsonschema)
   - [UISchemaElement](#uischemaelement)
   - [UISchemaRule та UISchemaCondition](#uischemarule-та-uischemacondition)
   - [FormOptions](#formoptions)
   - [FieldTags](#fieldtags)
   - [Options](#options)
   - [AccessLevel та FieldPermissions](#accesslevel-та-fieldpermissions)
   - [Translator та MapTranslator](#translator-та-maptranslator)
4. [Пакет `parser` — генерація схем](#пакет-parser)
   - [Генерація з Go-структур](#генерація-з-go-структур)
   - [Генерація з JSON-об'єктів](#генерація-з-json-обєктів)
   - [Генерація з OpenAPI 3.x](#генерація-з-openapi-3x)
5. [Пакет `api` — HTTP API](#пакет-api)
   - [Registry](#registry)
   - [Handler](#handler)
   - [HTTP ендпоінт](#http-ендпоінт)
6. [Struct Tags — повний довідник](#struct-tags)
   - [Базові теги](#базові-теги)
   - [Тег `form` — UI опції](#тег-form)
   - [Теги умовних правил](#теги-умовних-правил)
   - [i18n та renderer](#i18n-та-renderer)
7. [Маппінг типів Go → JSON Schema](#маппінг-типів)
8. [Приклади використання](#приклади-використання)
   - [Базова генерація](#базова-генерація)
   - [i18n — локалізація](#i18n--локалізація)
   - [Кастомні рендерери](#кастомні-рендерери)
   - [Ролі та дозволи](#ролі-та-дозволи)
   - [Категоризація (вкладки)](#категоризація-вкладки)
   - [Умовна видимість](#умовна-видимість)
   - [JSON Schema Draft 2019-09](#json-schema-draft-2019-09)
   - [OmitEmpty — фільтрація порожніх полів](#omitempty--фільтрація-порожніх-полів)
   - [OpenAPI 3.x → JSON Forms](#openapi-3x--json-forms)
   - [Генерація з JSON-даних](#генерація-з-json-даних)
   - [HTTP-сервер з реєстром типів](#http-сервер-з-реєстром-типів)
   - [Комбінований приклад](#комбінований-приклад)
9. [HTTP API — довідник](#http-api--довідник)
10. [Продуктивність](#продуктивність)
11. [Розробка](#розробка)

---

## Встановлення

```bash
go get github.com/holdemlab/ui-json-schema
```

Імпорт пакетів:

```go
import (
    "github.com/holdemlab/ui-json-schema/schema"  // типи, опції, i18n
    "github.com/holdemlab/ui-json-schema/parser"  // генерація схем
    "github.com/holdemlab/ui-json-schema/api"     // HTTP API, реєстр
)
```

---

## Архітектура

```
┌─────────────────────────────────────────────────────┐
│                  cmd/server/main.go                 │
│           HTTP-сервер (порт з $ADDR або :8080)      │
└──────────────────────┬──────────────────────────────┘
                       │ використовує
┌──────────────────────▼──────────────────────────────┐
│                    api/                              │
│  handler.go  — POST /schema/generate                │
│  registry.go — потоко-безпечний реєстр типів        │
└──────────────────────┬──────────────────────────────┘
                       │ викликає
┌──────────────────────▼──────────────────────────────┐
│                   parser/                           │
│  struct_parser.go  — Go struct → Schema             │
│  json_parser.go    — []byte JSON → Schema           │
│  openapi_parser.go — OpenAPI 3.x → Schema           │
└──────────────────────┬──────────────────────────────┘
                       │ використовує
┌──────────────────────▼──────────────────────────────┐
│                   schema/                           │
│  jsonschema.go — тип JSONSchema                     │
│  uischema.go   — UISchemaElement, правила, форми    │
│  tags.go       — FieldTags, парсинг struct-тегів    │
│  i18n.go       — Translator, MapTranslator          │
│  options.go    — Options, AccessLevel, Permissions  │
└─────────────────────────────────────────────────────┘
```

Бібліотека складається з трьох основних шарів:

1. **`schema`** — визначає типи даних (JSON Schema, UI Schema), конфігурацію та допоміжні інтерфейси.
2. **`parser`** — містить логіку генерації схем з різних джерел (Go struct, JSON, OpenAPI).
3. **`api`** — надає HTTP-обробник та реєстр типів для використання як мікросервісу.

---

## Пакет `schema`

### JSONSchema

Представляє документ JSON Schema (сумісний з Draft 7 та Draft 2019-09).

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
}
```

**Поля:**

| Поле | Тип | Опис |
|------|-----|------|
| `Schema` | `string` | URL стандарту JSON Schema (`$schema`) |
| `Type` | `string` | Тип значення: `"object"`, `"string"`, `"integer"`, `"number"`, `"boolean"`, `"array"`, `"null"` |
| `Properties` | `map[string]*JSONSchema` | Властивості об'єкта (для `type: "object"`) |
| `Items` | `*JSONSchema` | Схема елементів масиву (для `type: "array"`) |
| `AdditionalProperties` | `*JSONSchema` | Схема додаткових властивостей (для `map[string]T`) |
| `Required` | `[]string` | Список обов'язкових полів |
| `Format` | `string` | Формат значення: `"email"`, `"date-time"`, `"uri"`, тощо |
| `Default` | `any` | Значення за замовчуванням |
| `Enum` | `[]any` | Список допустимих значень |
| `Description` | `string` | Опис поля |
| `Title` | `string` | Заголовок поля |
| `Const` | `any` | Фіксоване значення |

**Конструктор:**

```go
func NewJSONSchema() *JSONSchema
```

Створює кореневий об'єкт JSON Schema з `$schema` встановленим на Draft 7 URL та `type` встановленим на `"object"`.

---

### UISchemaElement

Представляє елемент UI Schema для [JSON Forms](https://jsonforms.io/). Може бути лейаутом або контролом.

```go
type UISchemaElement struct {
    Type     string             `json:"type"`
    Label    string             `json:"label,omitempty"`
    Scope    string             `json:"scope,omitempty"`
    Elements []*UISchemaElement `json:"elements,omitempty"`
    Options  map[string]any     `json:"options,omitempty"`
    Rule     *UISchemaRule      `json:"rule,omitempty"`
}
```

**Поля:**

| Поле | Тип | Опис |
|------|-----|------|
| `Type` | `string` | Тип елемента: `"VerticalLayout"`, `"HorizontalLayout"`, `"Group"`, `"Categorization"`, `"Category"`, `"Control"` |
| `Label` | `string` | Мітка (для `Group`, `Category`, `Control`) |
| `Scope` | `string` | JSON Pointer шлях до властивості (для `Control`), наприклад `"#/properties/name"` |
| `Elements` | `[]*UISchemaElement` | Дочірні елементи (для лейаутів) |
| `Options` | `map[string]any` | Додаткові опції (`readonly`, `multi`, `renderer`) |
| `Rule` | `*UISchemaRule` | Умовне правило видимості/доступності |

**Типи лейаутів:**

| Тип | Опис | Конструктор |
|-----|------|-------------|
| `VerticalLayout` | Вертикальне розташування елементів (за замовчуванням) | `NewVerticalLayout()` |
| `HorizontalLayout` | Горизонтальне розташування | `NewHorizontalLayout()` |
| `Group` | Група з заголовком | `NewGroup(label string)` |
| `Categorization` | Кореневий контейнер для вкладок | `NewCategorization()` |
| `Category` | Вкладка (дочірній елемент `Categorization`) | `NewCategory(label string)` |

**Контрол:**

```go
func NewControl(scope string) *UISchemaElement
```

Створює контрол, що вказує на JSON Schema властивість через `scope`.

---

### UISchemaRule та UISchemaCondition

Правила дозволяють динамічно показувати/приховувати/активувати/деактивувати елементи UI на основі значень інших полів.

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

**Ефекти (константи):**

| Константа | Значення | Опис |
|-----------|----------|------|
| `EffectShow` | `"SHOW"` | Показати елемент, коли умова виконана |
| `EffectHide` | `"HIDE"` | Приховати елемент |
| `EffectEnable` | `"ENABLE"` | Активувати елемент |
| `EffectDisable` | `"DISABLE"` | Деактивувати елемент |

**Парсинг правил:**

```go
func ParseRuleExpression(expr string, effect string) *UISchemaRule
```

Приймає вираз у форматі `"field=value"` та ефект, повертає `*UISchemaRule`. Значення автоматично приводиться до `bool`, `int`, `float64` або `string`.

**Приклад:**

```go
rule := schema.ParseRuleExpression("is_active=true", schema.EffectShow)
// rule.Effect = "SHOW"
// rule.Condition.Scope = "#/properties/is_active"
// rule.Condition.Schema.Const = true (bool)
```

---

### FormOptions

Зберігає розпарсені дані тегу `form`.

```go
type FormOptions struct {
    Label     string // Мітка контролу
    Hidden    bool   // Приховати з UI
    Readonly  bool   // Поле тільки для читання
    Multiline bool   // Багаторядкове текстове поле
    Category  string // Назва категорії (вкладки)
    Layout    string // Переоприділення лейауту ("horizontal")
}
```

**Парсинг:**

```go
func ParseFormTag(tag string) FormOptions
```

Розбирає значення тегу `form`. Директиви розділені `;`.

**Приклади тегів:**

```
form:"label=Повне ім'я"
form:"hidden"
form:"readonly"
form:"multiline"
form:"category=Особисті дані"
form:"label=Ім'я;readonly;category=Профіль"
```

---

### FieldTags

Зберігає всі розпарсені struct-теги для одного поля.

```go
type FieldTags struct {
    Required  bool
    Default   any
    Enum      []any
    Format    string
    Form      string // сирий рядок тегу form
    I18nKey   string // ключ перекладу
    VisibleIf string // "field=value" для SHOW
    HideIf    string // "field=value" для HIDE
    EnableIf  string // "field=value" для ENABLE
    DisableIf string // "field=value" для DISABLE
    Renderer  string // назва кастомного рендерера
}
```

**Парсинг:**

```go
func ParseFieldTags(field reflect.StructField) FieldTags
```

Витягує всі schema-релевантні теги з поля. Значення `default` автоматично приводиться до типу поля (`bool`, `int`, `uint`, `float`, `string`). Значення `enum` розділяється комами.

---

### Options

Конфігурує поведінку генерації схем.

```go
type Options struct {
    Translator      Translator            // інтерфейс для локалізації
    Locale          string                // локаль: "uk", "en", тощо
    Draft           string                // "draft-07" (за замовчуванням) або "2019-09"
    Renderers       map[string]string     // scope → renderer name
    RolePermissions map[string]FieldPermissions // роль → дозволи полів
    Role            string                // активна роль
    OmitEmpty       bool                  // виключити omitempty-поля з нульовими значеннями
}
```

**Методи:**

```go
// DraftURL повертає URL $schema для обраного Draft
func (o Options) DraftURL() string

// DefaultOptions повертає Options з Draft = "draft-07"
func DefaultOptions() Options
```

**Значення `Draft`:**

| Значення | URL `$schema` |
|----------|---------------|
| `"draft-07"` (за замовчуванням) | `http://json-schema.org/draft-07/schema#` |
| `"2019-09"` | `https://json-schema.org/draft/2019-09/schema` |

**Поле `OmitEmpty`:**

Коли `OmitEmpty: true`, поля з тегом `json:",omitempty"` виключаються з генерованої JSON Schema та UI Schema, якщо відповідне значення є нульовим (zero value) для свого типу. За замовчуванням `false` — всі поля завжди включаються. Працює рекурсивно для вкладених структур.

---

### AccessLevel та FieldPermissions

Рівні доступу для ролевого керування полями.

```go
type AccessLevel int

const (
    AccessFull     AccessLevel = iota // 0 — повний доступ (за замовчуванням)
    AccessReadOnly                    // 1 — тільки читання
    AccessHidden                      // 2 — поле приховане
)

type FieldPermissions map[string]AccessLevel
```

**Використання:**

```go
opts := schema.Options{
    Role: "viewer",
    RolePermissions: map[string]schema.FieldPermissions{
        "viewer": {
            "name":  schema.AccessReadOnly, // readonly в UI
            "role":  schema.AccessHidden,   // прихований з UI
        },
        "admin": {
            // всі поля доступні
        },
    },
}
```

---

### Translator та MapTranslator

Інтерфейс для локалізації міток.

```go
type Translator interface {
    Translate(key, locale string) string
}
```

`Translate` повертає локалізований рядок за ключем та локаллю. Якщо переклад не знайдено, повертає ключ без змін.

**MapTranslator** — реалізація на основі вкладених map:

```go
func NewMapTranslator(m map[string]map[string]string) *MapTranslator
```

Зовнішній ключ — локаль, внутрішній — ключ перекладу:

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

Ви можете створити власну реалізацію `Translator`, наприклад, з підтримкою файлів `.po`, бази даних або зовнішнього сервісу.

---

## Пакет `parser`

### Генерація з Go-структур

#### GenerateJSONSchema

```go
func GenerateJSONSchema(v any) (*schema.JSONSchema, error)
```

Генерує JSON Schema (Draft 7) з Go-значення. Приймає struct або вказівник на struct.

#### GenerateJSONSchemaWithOptions

```go
func GenerateJSONSchemaWithOptions(v any, opts schema.Options) (*schema.JSONSchema, error)
```

Те ж саме, з можливістю задати опції (Draft, Translator тощо).

#### GenerateUISchema

```go
func GenerateUISchema(v any) (*schema.UISchemaElement, error)
```

Генерує UI Schema (JSON Forms) з Go-значення.

#### GenerateUISchemaWithOptions

```go
func GenerateUISchemaWithOptions(v any, opts schema.Options) (*schema.UISchemaElement, error)
```

Те ж саме, з можливістю задати опції (i18n, renderers, permissions).

**Логіка генерації:**

1. Ітерує всі експортовані поля структури.
2. Поля з тегом `json:"-"` пропускаються.
3. Якщо `OmitEmpty: true` — поля з `json:",omitempty"` та нульовим значенням пропускаються (рекурсивно для вкладених структур).
4. Вкладені структури (крім `time.Time`) обробляються рекурсивно як `"object"` / `Group`.
5. `time.Time` маппиться як `"string"` з `format: "date-time"`.
6. Вказівники (`*T`) розгортаються до базового типу.
7. Для UI Schema:
   - `form:"hidden"` або `AccessHidden` → поле виключається.
   - `AccessReadOnly` → `options.readonly = true`.
   - Вкладені структури → `Group` з назвою поля як мітка.
   - Категорії → автоматична обгортка в `Categorization` → `Category`.
   - Правила застосовуються за пріоритетом: `visibleIf` → `hideIf` → `enableIf` → `disableIf`.

---

### Генерація з JSON-об'єктів

#### GenerateFromJSON

```go
func GenerateFromJSON(data []byte) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Генерує обидві схеми з сирого JSON. Вхідні дані мають бути JSON-об'єктом. Всі поля вважаються необов'язковими.

#### GenerateFromJSONWithOptions

```go
func GenerateFromJSONWithOptions(data []byte, opts schema.Options) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Те ж саме з опціями.

**Сигнальні помилки:**

| Змінна | Опис |
|--------|------|
| `ErrInvalidJSON` | Вхідні дані не є валідним JSON |
| `ErrNotAnObject` | Верхній рівень JSON не є об'єктом |

**Автовизначення типів JSON → JSON Schema:**

| JSON-значення | JSON Schema `type` |
|---------------|---------------------|
| `null` | `"null"` |
| `true` / `false` | `"boolean"` |
| `42` (ціле число) | `"integer"` |
| `3.14` (дробове) | `"number"` |
| `"text"` | `"string"` |
| `[...]` | `"array"` (тип items з першого елементу) |
| `{...}` | `"object"` (properties рекурсивно) |

---

### Генерація з OpenAPI 3.x

#### GenerateFromOpenAPI

```go
func GenerateFromOpenAPI(data []byte, schemaName string) (*schema.JSONSchema, *schema.UISchemaElement, error)
```

Парсить OpenAPI 3.x JSON-документ та генерує обидві схеми для зазначеного компонента. `schemaName` має відповідати ключу в `components.schemas`.

**Сигнальні помилки:**

| Змінна | Опис |
|--------|------|
| `ErrInvalidOpenAPI` | Документ не є валідним OpenAPI |
| `ErrSchemaNotFound` | Зазначена схема не знайдена в `components.schemas` |

**Особливості:**

- Підтримує `$ref` посилання всередині документа (формат: `#/components/schemas/Name`).
- Конвертує `type`, `properties`, `items`, `required`, `format`, `default`, `enum`, `description`, `title`.
- Вкладені об'єкти з `properties` стають `Group` в UI Schema.

---

## Пакет `api`

### Registry

Потоко-безпечний реєстр Go-типів для генерації схем за назвою.

```go
type Registry struct { /* unexported fields */ }

func NewRegistry() *Registry
func (r *Registry) Register(name string, v any)
func (r *Registry) Lookup(name string) (any, error)
func (r *Registry) Names() []string
```

| Метод | Опис |
|-------|------|
| `NewRegistry()` | Створює порожній реєстр |
| `Register(name, v)` | Реєструє екземпляр struct під ім'ям. Перезаписує при повторі. |
| `Lookup(name)` | Повертає зареєстрований екземпляр або помилку |
| `Names()` | Повертає список всіх зареєстрованих імен |

**Приклад:**

```go
reg := api.NewRegistry()
reg.Register("User", User{})
reg.Register("Order", Order{})

names := reg.Names() // ["User", "Order"]
v, err := reg.Lookup("User") // User{}, nil
```

---

### Handler

HTTP-обробник для генерації схем.

```go
type Handler struct { /* unexported fields */ }

func NewHandler(registry *Registry) *Handler
func (h *Handler) GenerateHandler(w http.ResponseWriter, r *http.Request)
```

**`MaxBodySize`** — обмеження розміру тіла запиту: `2 MB` (2 << 20).

---

### HTTP ендпоінт

**`POST /schema/generate`**

Приймає JSON-тіло з одним з двох полів:

#### Варіант 1: за зареєстрованим типом

```json
{"type": "User"}
```

Шукає `"User"` в реєстрі та генерує схеми з Go-структури.

#### Варіант 2: з сирого JSON

```json
{"data": {"name": "John", "age": 30}}
```

Генерує схеми з переданого JSON-об'єкта.

#### Відповідь (200 OK)

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

#### Помилки

| HTTP код | Причина |
|----------|---------|
| 405 | Не POST метод |
| 400 | Некоректний JSON, відсутні `type` і `data` |
| 404 | Тип не знайдено в реєстрі |
| 500 | Внутрішня помилка генерації |

Формат помилки:

```json
{"error": "type 'Unknown' not found"}
```

---

## Struct Tags

### Базові теги

| Тег | Значення | Вплив на JSON Schema | Приклад |
|-----|----------|---------------------|---------|
| `json:"name"` | Ім'я поля | Ключ в `properties` | `json:"email"` |
| `json:"-"` | Пропустити | Поле виключається повністю | `json:"-"` |
| `json:",omitempty"` | Пропустити якщо порожнє | З `OmitEmpty: true` поле виключається при zero value | `json:"notes,omitempty"` |
| `required:"true"` | Обов'язкове | Додається до `required[]` | `required:"true"` |
| `default:"value"` | За замовчуванням | Встановлює `default` (авто-приведення типу) | `default:"true"`, `default:"42"` |
| `enum:"a,b,c"` | Допустимі значення | Встановлює `enum[]` | `enum:"admin,user,moderator"` |
| `format:"fmt"` | Формат | Встановлює `format` | `format:"email"`, `format:"date-time"`, `format:"uri"` |

**Приведення типу `default`:**

| Тип поля Go | Приведення |
|-------------|-----------|
| `bool` | `"true"` → `true`, `"false"` → `false` |
| `int`, `int8`–`int64` | `"42"` → `42` |
| `uint`, `uint8`–`uint64` | `"10"` → `10` |
| `float32`, `float64` | `"3.14"` → `3.14` |
| `string` | Без змін |

### Тег `form` — UI опції

Директиви розділяються символом `;`:

| Директива | Опис | Приклад |
|-----------|------|---------|
| `label=Текст` | Мітка контролу | `form:"label=Повне ім'я"` |
| `hidden` | Приховати з UI Schema | `form:"hidden"` |
| `readonly` | Тільки для читання | `form:"readonly"` |
| `multiline` | Багаторядкове поле | `form:"multiline"` |
| `category=Назва` | Категорія (вкладка) | `form:"category=Особисті"` |
| `layout=horizontal` | Переоприділення лейауту | `form:"layout=horizontal"` |

**Комбінації:**

```go
type Profile struct {
    Name  string `json:"name" form:"label=Ім'я;category=Основне"`
    Bio   string `json:"bio" form:"multiline;readonly;category=Додатково"`
    Token string `json:"token" form:"hidden"`
}
```

### Теги умовних правил

Формат значення: `"поле=значення"`.

| Тег | Ефект | Приклад | Результат |
|-----|-------|---------|-----------|
| `visibleIf:"field=val"` | `SHOW` | `visibleIf:"is_active=true"` | Показувати, коли `is_active == true` |
| `hideIf:"field=val"` | `HIDE` | `hideIf:"role=admin"` | Приховати, коли `role == "admin"` |
| `enableIf:"field=val"` | `ENABLE` | `enableIf:"agreed=true"` | Активувати, коли `agreed == true` |
| `disableIf:"field=val"` | `DISABLE` | `disableIf:"locked=true"` | Деактивувати, коли `locked == true` |

**Пріоритет правил** (застосовується лише перше знайдене):

`visibleIf` → `hideIf` → `enableIf` → `disableIf`

**Автоприведення значень:**

- `"true"` / `"false"` → `bool`
- `"42"` → `int`
- `"3.14"` → `float64`
- Все інше → `string`

**Результат у JSON:**

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

### i18n та renderer

| Тег | Опис | Приклад |
|-----|------|---------|
| `i18n:"key"` | Ключ перекладу для мітки | `i18n:"user.name"` |
| `renderer:"name"` | Кастомний рендерер | `renderer:"color-picker"` |

Якщо `Translator` задано в `Options` та є переклад для `i18n`-ключа — мітка контролу замінюється на переклад. Якщо перекладу нема — використовується `form:"label=..."` або Go-назва поля.

Рендерер з тегу має пріоритет над `Options.Renderers`.

---

## Маппінг типів

| Go тип | JSON Schema `type` | Додатково |
|--------|---------------------|-----------|
| `string` | `"string"` | — |
| `bool` | `"boolean"` | — |
| `int`, `int8`, `int16`, `int32`, `int64` | `"integer"` | — |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `"integer"` | — |
| `float32`, `float64` | `"number"` | — |
| `time.Time` | `"string"` | `format: "date-time"` |
| `[]T`, `[N]T` | `"array"` | `items` — схема `T` |
| `map[string]T` | `"object"` | `additionalProperties` — схема `T` |
| `map[K]T` (K ≠ string) | `"object"` | Без `additionalProperties` |
| вкладений `struct` | `"object"` | `properties` рекурсивно |
| `*T` (вказівник) | розгортається до `T` | — |
| інші типи | `"string"` | Fallback |

---

## Приклади використання

### Базова генерація

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

**Результат JSON Schema:**

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

**Результат UI Schema:**

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

> Зверніть увагу: поле `id` відсутнє в UI Schema через `form:"hidden"`.

---

### i18n — локалізація

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
    // Контрол "name"  → label: "Ім'я"
    // Контрол "email" → label: "Електронна пошта"
}
```

**Власна реалізація Translator:**

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
        return key // fallback до ключа
    }
    return val
}
```

---

### Кастомні рендерери

Рендерери дозволяють вказати JSON Forms, який UI-компонент використовувати для поля.

**Через struct tag (пріоритет):**

```go
type Config struct {
    Color  string `json:"color" renderer:"color-picker"`
    Rating int    `json:"rating" renderer:"star-rating"`
}

ui, _ := parser.GenerateUISchema(Config{})
```

Результат:

```json
{
  "type": "Control",
  "scope": "#/properties/color",
  "options": { "renderer": "color-picker" }
}
```

**Через Options (за scope):**

```go
opts := schema.Options{
    Renderers: map[string]string{
        "#/properties/rating": "star-rating",
        "#/properties/avatar": "image-upload",
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Config{}, opts)
```

> Тег `renderer` має пріоритет над `Options.Renderers`.

---

### Ролі та дозволи

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
            "status":  schema.AccessReadOnly,  // можна бачити, не можна міняти
            "author":  schema.AccessHidden,     // зовсім не бачить
        },
        "admin": {
            // всі поля з повним доступом (AccessFull за замовчуванням)
        },
    },
}

ui, _ := parser.GenerateUISchemaWithOptions(Article{}, opts)
// title   → звичайний контрол
// content → звичайний контрол
// status  → readonly контрол
// author  → відсутній в UI Schema
```

---

### Категоризація (вкладки)

Коли хоча б одне поле має тег `form:"category=..."`, кореневий лейаут автоматично стає `Categorization`, а поля групуються у відповідні `Category`.

```go
type RegistrationForm struct {
    FirstName string `json:"first_name" form:"label=Ім'я;category=Особисті дані"`
    LastName  string `json:"last_name" form:"label=Прізвище;category=Особисті дані"`
    Email     string `json:"email" form:"category=Контакти" format:"email"`
    Phone     string `json:"phone" form:"category=Контакти"`
    Company   string `json:"company" form:"category=Робота"`
    Position  string `json:"position" form:"category=Робота"`
}

ui, _ := parser.GenerateUISchema(RegistrationForm{})
```

Результат:

```json
{
  "type": "Categorization",
  "elements": [
    {
      "type": "Category",
      "label": "Особисті дані",
      "elements": [
        { "type": "Control", "scope": "#/properties/first_name", "label": "Ім'я" },
        { "type": "Control", "scope": "#/properties/last_name", "label": "Прізвище" }
      ]
    },
    {
      "type": "Category",
      "label": "Контакти",
      "elements": [
        { "type": "Control", "scope": "#/properties/email" },
        { "type": "Control", "scope": "#/properties/phone" }
      ]
    },
    {
      "type": "Category",
      "label": "Робота",
      "elements": [
        { "type": "Control", "scope": "#/properties/company" },
        { "type": "Control", "scope": "#/properties/position" }
      ]
    }
  ]
}
```

---

### Умовна видимість

```go
type Survey struct {
    HasPet    bool   `json:"has_pet"`
    PetName   string `json:"pet_name" visibleIf:"has_pet=true" form:"label=Ім'я тварини"`
    PetAge    int    `json:"pet_age" visibleIf:"has_pet=true"`
    Country   string `json:"country"`
    State     string `json:"state" enableIf:"country=US" form:"label=Штат"`
    IsMinor   bool   `json:"is_minor"`
    ParentName string `json:"parent_name" visibleIf:"is_minor=true"`
    Reason    string `json:"reason" hideIf:"has_pet=false"`
}

ui, _ := parser.GenerateUISchema(Survey{})
```

Контрол `pet_name` в UI Schema:

```json
{
  "type": "Control",
  "scope": "#/properties/pet_name",
  "label": "Ім'я тварини",
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

### JSON Schema Draft 2019-09

```go
opts := schema.Options{Draft: "2019-09"}

jsonSchema, _ := parser.GenerateJSONSchemaWithOptions(User{}, opts)
// $schema → "https://json-schema.org/draft/2019-09/schema"
```

---

### OmitEmpty — фільтрація порожніх полів

Коли `OmitEmpty: true`, поля з тегом `json:",omitempty"` виключаються зі схеми, якщо їх значення є нульовим (zero value) для свого типу. Це працює для обох схем — JSON Schema та UI Schema, а також рекурсивно для вкладених структур.

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

    // Порожня структура — omitempty поля виключаються
    empty := Article{Title: "Hello"}
    s1, _ := parser.GenerateJSONSchemaWithOptions(empty, opts)
    b1, _ := json.MarshalIndent(s1, "", "  ")
    fmt.Println("Empty:", string(b1))
    // properties: title, content (без notes, tags, views)

    // Заповнена структура — omitempty поля включаються
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
    // properties: title, content, notes, tags, views (всі присутні)
}
```

**Результат для порожньої структури:**

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

> Поля `notes`, `tags` та `views` відсутні, бо мають `omitempty` і нульові значення.

**Нульові значення за типами:**

| Тип | Нульове значення |
|-----|------------------|
| `string` | `""` |
| `int`, `float` тощо | `0` |
| `bool` | `false` |
| `slice`, `map` | `nil` |
| `*T` (вказівник) | `nil` |
| `struct` | всі поля нульові |

> **Примітка:** без `OmitEmpty: true` (за замовчуванням) всі поля з `omitempty` завжди включаються в схему.

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
                        "description": "Ім'я тваринки"
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
// $ref "#/components/schemas/Owner" розрезолвиться автоматично
```

---

### Генерація з JSON-даних

```go
data := []byte(`{
    "name": "Тарас",
    "age": 30,
    "is_active": true,
    "scores": [95, 87, 92],
    "address": {
        "city": "Київ",
        "street": "Хрещатик",
        "zip": "01001"
    }
}`)

jsonSchema, uiSchema, err := parser.GenerateFromJSON(data)
if err != nil {
    // err == parser.ErrInvalidJSON або parser.ErrNotAnObject
    panic(err)
}
```

Результат — JSON Schema з автоматично визначеними типами:
- `"name"` → `string`
- `"age"` → `integer` (ціле число)
- `"is_active"` → `boolean`
- `"scores"` → `array` з `items: {type: "integer"}`
- `"address"` → `object` з `properties`

---

### HTTP-сервер з реєстром типів

```go
package main

import (
    "log"
    "net/http"

    handler "github.com/holdemlab/ui-json-schema/api"
)

type User struct {
    Name     string `json:"name" required:"true" form:"label=Ім'я"`
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

**Запити:**

```bash
# Генерація з типу
curl -s -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"type": "User"}' | jq .

# Генерація з JSON
curl -s -X POST http://localhost:8080/schema/generate \
  -H "Content-Type: application/json" \
  -d '{"data": {"title": "Laptop", "price": 999.99}}' | jq .
```

---

### Комбінований приклад

Повний приклад з усіма можливостями:

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/holdemlab/ui-json-schema/parser"
    "github.com/holdemlab/ui-json-schema/schema"
)

type Employee struct {
    // Основне
    ID        int    `json:"id" form:"hidden"`
    FirstName string `json:"first_name" required:"true" form:"label=Ім'я;category=Основне" i18n:"emp.first_name"`
    LastName  string `json:"last_name" required:"true" form:"label=Прізвище;category=Основне" i18n:"emp.last_name"`
    Email     string `json:"email" required:"true" format:"email" form:"category=Основне"`

    // Робота
    Department string `json:"department" enum:"engineering,marketing,sales,hr" form:"category=Робота"`
    Position   string `json:"position" form:"category=Робота"`
    Salary     int    `json:"salary" form:"category=Робота;readonly"`

    // Додатково
    Bio       string `json:"bio" form:"multiline;category=Додатково"`
    IsRemote  bool   `json:"is_remote" default:"false" form:"category=Додатково"`
    Office    string `json:"office" hideIf:"is_remote=true" form:"category=Додатково"`
    Equipment string `json:"equipment" visibleIf:"is_remote=true" form:"category=Додатково" renderer:"equipment-selector"`
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
                // повний доступ до всіх полів
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

## HTTP API — довідник

### Запуск вбудованого сервера

```bash
# За замовчуванням :8080
make run

# Або з кастомним портом
ADDR=":3000" go run cmd/server/main.go
```

### Ендпоінт

| Метод | Шлях | Опис |
|-------|------|------|
| `POST` | `/schema/generate` | Генерація JSON Schema + UI Schema |

### Формат запиту

**Content-Type:** `application/json`  
**Максимальний розмір тіла:** 2 MB

**Поля (одне з двох):**

| Поле | Тип | Опис |
|------|-----|------|
| `type` | `string` | Ім'я зареєстрованого Go-типу |
| `data` | `object` | Сирий JSON-об'єкт для аналізу |

### Формат відповіді

**Успіх (200):**

```json
{
  "schema":   { /* JSONSchema */ },
  "uischema": { /* UISchemaElement */ }
}
```

**Помилка (4xx/5xx):**

```json
{
  "error": "повідомлення про помилку"
}
```

### Коди помилок

| Код | Причина |
|-----|---------|
| `405 Method Not Allowed` | Використано не POST метод |
| `400 Bad Request` | Некоректний JSON або відсутні обов'язкові поля |
| `404 Not Found` | Тип не знайдено в реєстрі |
| `500 Internal Server Error` | Помилка генерації схем |

---

## Продуктивність

Бенчмарки на Intel i7-14700HX:

| Операція | Час | Алокації |
|----------|-----|----------|
| JSON Schema з малої структури (5 полів) | ~2.2 µs | 9 allocs |
| JSON Schema з середньої структури (15 полів) | ~10.4 µs | 42 allocs |
| JSON Schema з великої структури (40+ полів) | ~25.5 µs | 93 allocs |
| Генерація з 1 MB JSON | ~3.9 ms | 3,536 allocs |
| Генерація з 2 MB JSON | ~6.3 ms | 3,536 allocs |

Всі операції значно нижче цільового показника 100 мс для JSON до 2 MB.

---

## Розробка

### Вимоги

- Go 1.24+
- golangci-lint v2
- GNU Make

### Make-команди

| Команда | Опис |
|---------|------|
| `make test` | Запуск тестів |
| `make test-cover` | Тести з покриттям (мінімум 80%) |
| `make bench` | Бенчмарки |
| `make lint` | Запуск лінтера |
| `make build` | Збірка серверу |
| `make run` | Запуск серверу |

### CI Pipeline

1. **Lint** — golangci-lint v2 з конфігурацією `.golangci.yml`
2. **Test** — всі тести + перевірка покриття ≥ 80%
3. **Build** — `go build ./...`
4. **Tag** — автоматичне створення тегу (тільки на `master`) за Conventional Commits

### Conventional Commits

| Формат | Версія |
|--------|--------|
| `BREAKING CHANGE:` в тілі | Major (v1.0.0 → v2.0.0) |
| `feat: ...` | Minor (v0.1.0 → v0.2.0) |
| `fix: ...`, `docs: ...`, `chore: ...` | Patch (v0.1.0 → v0.1.1) |

---

## Ліцензія

MIT
