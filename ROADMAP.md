# ROADMAP — JSON Schema & UI Schema Generator

Автоматичне генерування JSON Schema та UI Schema з Go-структур і JSON-об'єктів для JSON Forms.

---

## Етап 0 — Ініціалізація проєкту ✅

- [x] Ініціалізація Go-модуля (`go mod init`)
- [x] Налаштування структури каталогів (`schema/`, `parser/`, `api/`, `cmd/server/`)
- [x] Налаштування GolangCI-Lint v2 (базовий набір лінтерів)
- [x] Налаштування CI (GitHub Actions: lint + test + build)
- [x] Створення `Makefile` (lint, test, test-cover, build, fmt, clean)
- [x] Створення `.gitignore`

**Результат:** порожній проєкт із робочим CI-пайплайном.

---

## Етап 1 — Ядро: генерація JSON Schema з Go-структур ✅

- [x] `schema/jsonschema.go` — базові типи JSON Schema (Draft 7)
- [x] `parser/struct_parser.go` — аналіз Go struct через `reflect`
- [x] Підтримка примітивних типів: `string`, `int`, `int32`, `int64`, `float32`, `float64`, `bool`
- [x] Підтримка `time.Time` → `{"type":"string","format":"date-time"}`
- [x] Підтримка вкладених `struct`
- [x] Підтримка `slices` (`[]T`) → `{"type":"array","items":{…}}`
- [x] Підтримка `map[string]T` → `{"type":"object","additionalProperties":{…}}`
- [x] Читання тегу `json` для імені поля
- [x] Unit-тести (покриття 91.3% ≥ 80%)

**Результат:** `GenerateJSONSchema(v any) (JSONSchema, error)` — генерує коректну JSON Schema з довільної Go-структури.

---

## Етап 2 — Підтримка struct-тегів ✅

- [x] `schema/tags.go` — парсинг кастомних тегів
- [x] Тег `required:"true"` → поле додається до `required`
- [x] Тег `default:"…"` → `"default": …`
- [x] Тег `enum:"a,b,c"` → `"enum": ["a","b","c"]`
- [x] Тег `format:"email"` / `format:"date"` / `format:"date-time"` → `"format": "…"`
- [x] Обробка `json:"-"` (пропуск поля)
- [x] Обробка `json:",omitempty"`
- [x] Unit-тести на кожен тег (покриття 91.8%)

**Результат:** JSON Schema враховує всі задекларовані теги.

---

## Етап 3 — Генерація UI Schema ✅

- [x] `schema/uischema.go` — типи UI Schema (JSON Forms)
- [x] Автоматичне створення `VerticalLayout` з `Control`-елементами
- [x] `scope` → `#/properties/<field>`
- [x] Парсинг тегу `form:"…"`:
  - `label=Full name` → `"label": "Full name"`
  - `hidden` → елемент не додається до UI Schema
  - `readonly` → `"options": {"readonly": true}`
  - `multiline` → `"options": {"multi": true}`
- [x] Рекурсивна обробка вкладених структур (Group / nested layout)
- [x] Unit-тести (покриття 92.4%)

**Результат:** `GenerateUISchema(v any) (UISchema, error)` — генерує UI Schema, сумісну з JSON Forms.

---

## Етап 4 — Rules (умовна логіка)

- [x] Парсинг тегів `visibleIf`, `hideIf`, `enableIf`, `disableIf`
- [x] Генерація `rule` блоку в UI Schema:
  - `effect`: `SHOW` / `HIDE` / `ENABLE` / `DISABLE`
  - `condition`: `scope` + `schema.const`
- [x] Підтримка різних типів значень у condition (`bool`, `string`, `int`, `float`)
- [x] Пріоритет правил: `visibleIf` → `hideIf` → `enableIf` → `disableIf`
- [x] Інтеграція в `buildUIElements` через `applyRule()`
- [x] Unit-тести (покриття 93.6%)

**Результат:** UI Schema з умовними правилами відображення полів.

---

## Етап 5 — Генерація JSON Schema з довільного JSON

- [x] `parser/json_parser.go` — парсинг `[]byte` → `map[string]any`
- [x] Визначення типів значень (`string`, `number`, `integer`, `boolean`, `null`)
- [x] Розрізнення `integer` vs `number` (через `math.Trunc`)
- [x] Вкладені об'єкти → вкладені `properties` + Group у UI Schema
- [x] Масиви → `items` (тип визначається з першого елемента)
- [x] Порожні масиви → порожній `items` schema
- [x] Масиви об'єктів → `items.properties` з першого елемента
- [x] Генерація UI Schema з JSON-об'єкта (VerticalLayout, Controls, Groups)
- [x] Всі поля `optional` (без `required`)
- [x] Валідація: помилка для не-об'єктних JSON (масив, рядок, число, null)
- [x] Unit-тести (покриття 94.4%)

**Результат:** `GenerateFromJSON(data []byte) (*JSONSchema, *UISchemaElement, error)` — генерує обидві схеми з довільного JSON.

---

## Етап 6 — HTTP API

- [x] `api/registry.go` — реєстр Go-типів (`Registry`) з `Register`, `Lookup`, `Names`; thread-safe через `sync.RWMutex`
- [x] `api/handler.go` — HTTP-хендлер `POST /schema/generate` (`GenerateHandler`)
- [x] Прийом `{"type":"Name"}` → генерація з зареєстрованого Go-типу
- [x] Прийом `{"data":{…}}` → генерація з JSON-об'єкта
- [x] Пріоритет: `type` > `data` при наявності обох полів
- [x] Формат відповіді: `{"schema":{…},"uischema":{…}}`
- [x] Валідація: помилки для невалідного JSON, порожнього body, відсутнього type/data, невідомого типу, масиву замість об'єкта
- [x] Коректні HTTP статус-коди: 200 OK, 400 Bad Request, 404 Not Found, 405 Method Not Allowed
- [x] `cmd/server/main.go` — HTTP-сервер з `http.ListenAndServe`, конфігурація адреси через `ADDR` env
- [x] Ліміт body 2 МБ (`maxRequestBody`)
- [x] Інтеграційні тести для API (20+ тестів: registry + handler) — покриття API 93.9%
- [x] Unit-тести (загальне покриття 91.8%)

**Результат:** працюючий HTTP-сервер, що віддає JSON Schema + UI Schema.

---

## Етап 7 — Продуктивність та якість ✅

- [x] Бенчмарки генерації (JSON до 1–2 МБ < 100 мс)
  - Small struct (5 полів): ~2.2 µs
  - Medium struct (15+ полів): ~10.4 µs
  - Large struct (40+ полів): ~25.5 µs
  - 1 МБ JSON: ~3.9 мс ✅ (< 100 мс)
  - 2 МБ JSON: ~6.3 мс ✅ (< 100 мс)
- [x] Профілювання та оптимізація — не потрібна (всі бенчмарки значно нижче 100 мс)
- [x] Перевірка покриття тестами ≥ 80% — загальне покриття **91.8%** (parser 95.1%, schema 94.3%, API 93.9%)
- [x] Перевірка сумісності згенерованих схем із JSON Forms — `parser/compatibility_test.go` (3 тести: StructSchema, UISchema, FromJSON)
- [x] Фінальний прохід лінтером — **0 issues**
- [x] README з прикладами використання — `README.md` (features, installation, quick start, tags table, type mapping, project structure, benchmarks)
- [x] Makefile `bench` таргет для запуску бенчмарків

**Результат:** production-ready бібліотека з документацією.

---

## Етап 8 — Розширення ✅

- [x] i18n для labels — `schema.Translator` інтерфейс, `MapTranslator` реалізація, `i18n` struct tag, автоматичний переклад labels
- [x] Custom renderers mapping — `renderer` struct tag + `Options.Renderers` map (тег має пріоритет)
- [x] Permissions / readonly by role — `Options.Role` + `Options.RolePermissions`, рівні доступу: ReadWrite / ReadOnly / Hidden
- [x] OpenAPI → JSON Forms — `parser.GenerateFromOpenAPI()`, підтримка `$ref`, вкладених об'єктів, масивів, enum
- [x] Підтримка JSON Schema Draft 2019-09 — `Options.Draft`, `DraftURL()`, обидва парсери (struct + JSON)
- [x] Custom layouts (Categorization) — `form:"category=..."` тег, автоматичне групування в Categorization/Category, fallback "Other"

**Нові файли:**
- `schema/i18n.go` — Translator інтерфейс та MapTranslator
- `schema/options.go` — Options struct (Draft, Translator, Renderers, RolePermissions)
- `parser/openapi_parser.go` — OpenAPI 3.x → JSON Schema + UI Schema

**Покриття тестами:** 92.4% загальне (parser 93.9%, schema 95.4%, API 93.9%)
**Лінт:** 0 issues

---

## Етап 9 — Validation Constraints (JSON Schema) ✅

Додавання підтримки валідаційних обмежень JSON Schema через struct-теги.

- [x] Додати поля до `JSONSchema`: `MinLength`, `MaxLength`, `Minimum`, `Maximum`, `Pattern`, `Description`
- [x] Додати парсинг нових тегів у `ParseFieldTags`:
  - `minLength:"3"` → `"minLength": 3`
  - `maxLength:"100"` → `"maxLength": 100`
  - `minimum:"0"` → `"minimum": 0`
  - `maximum:"999"` → `"maximum": 999`
  - `pattern:"^[a-z]+$"` → `"pattern": "^[a-z]+$"`
  - `description:"Please enter your name"` → `"description": "Please enter your name"`
- [x] Застосування нових тегів у `applyTags()` (`parser/struct_parser.go`)
- [x] Підтримка цілих та дробових значень для `minimum`/`maximum`
- [x] Unit-тести на кожен новий тег + комбінації
- [x] Лінт: 0 issues

**Файли:** `schema/jsonschema.go`, `schema/tags.go`, `parser/struct_parser.go`

**Результат:** JSON Schema з валідаційними обмеженнями — `minLength`, `maxLength`, `minimum`, `maximum`, `pattern`, `description`.

---

## Етап 10 — HorizontalLayout ✅

Підтримка горизонтального лейауту через `form:"layout=horizontal"` тег. Сусідні поля з однаковим лейаутом групуються в один `HorizontalLayout`.

- [x] Реалізувати групування полів у `buildUIElements` за `FormOptions.Layout`:
  - Послідовні поля з `form:"layout=horizontal"` об'єднуються в `HorizontalLayout`
  - Поля без layout залишаються як окремі Control (VerticalLayout за замовчуванням)
  - Горизонтальне групування працює всередині Category, Group, кореневого VerticalLayout
- [x] ~~Аналогічна підтримка в `buildOpenAPIUISchema` (OpenAPI парсер)~~ — пропущено: OpenAPI-специфікації не несуть layout-хінтів
- [x] Unit-тести:
  - Групування 2+ полів у HorizontalLayout
  - Мікс: horizontal + vertical поля
  - HorizontalLayout всередині Category
  - HorizontalLayout всередині вкладеної структури (Group)
  - Одне поле з layout=horizontal → не створювати HorizontalLayout (залишити Control)
- [x] Лінт: 0 issues

**Файли:** `parser/struct_parser.go`, `parser/openapi_parser.go`

**Результат:** `form:"layout=horizontal"` групує сусідні поля в `HorizontalLayout`.

**Приклад:**
```go
type Person struct {
    FirstName string `json:"firstName" form:"layout=horizontal"`
    LastName  string `json:"lastName" form:"layout=horizontal"`
    Email     string `json:"email"`
}
```
Генерує:
```json
{
  "type": "VerticalLayout",
  "elements": [
    {
      "type": "HorizontalLayout",
      "elements": [
        { "type": "Control", "scope": "#/properties/firstName" },
        { "type": "Control", "scope": "#/properties/lastName" }
      ]
    },
    { "type": "Control", "scope": "#/properties/email" }
  ]
}
```

---

## Етап 11 — Rules та i18n на Layout-елементах ✅

Розширення Rules та i18n з рівня Control на рівень Category, Group та інших лейаутів.

### 11.1 — Rules на Category / Group ✅

- [x] Додати тег `categoryRule:"visibleIf=field:value"` або розширити `form` тег для rules на category:
  - `form:"category=Address;visibleIf=provideAddress:true"` → Category "Address" має rule SHOW
- [x] Генерація `rule` блоку на `Category` елементі в `buildCategorization()`
- [x] Підтримка всіх ефектів: SHOW, HIDE, ENABLE, DISABLE
- [x] Unit-тести:
  - Rule на Category (SHOW/HIDE)
  - Category без rule (без регресії)
  - Кілька категорій — одна з rule, інша без

### 11.2 — i18n на Category ✅

- [x] Додати підтримку `i18n` ключа на Category через розширений `form` тег:
  - `form:"category=Personal;i18n=category.personal"` → Category отримує i18n ключ
- [x] Переклад label категорії через `Translator` (аналогічно Control labels)
- [x] Додати поле `I18n` до `UISchemaElement` (`json:"i18n,omitempty"`)
- [x] Unit-тести:
  - Category з i18n ключем
  - Category без i18n (fallback на label)
  - Переклад label категорії через Translator

### 11.3 — Rules на вкладених структурах (Group) ✅

- [x] Підтримка `visibleIf`/`hideIf` на полі-структурі → rule застосовується до Group:
  ```go
  Address AddressStruct `json:"address" visibleIf:"provideAddress=true"`
  ```
- [x] Unit-тести

- [x] Лінт: 0 issues

**Файли:** `schema/uischema.go`, `schema/tags.go`, `parser/struct_parser.go`

**Результат:** Повна підтримка Rules та i18n на всіх рівнях UI Schema — Control, Group, Category.

---

## Етап 12 — Іменовані Layout-групи ✅

- [x] Підтримка іменованих layout-груп через `form:"layout=horizontal:groupName"`:
  - Несусідні поля з однаковою назвою групи об'єднуються в один `HorizontalLayout`
  - Дозволяє гнучке компонування без додавання вкладених структур
- [x] Парсинг `layout=horizontal:name` у `ParseFormTag` → зберігання імені групи у `FormOptions`
- [x] Оновлення `groupHorizontalElements` для підтримки іменованих груп
- [x] Unit-тести:
  - Несусідні поля з однаковою назвою групи → один HorizontalLayout
  - Різні назви груп → окремі HorizontalLayout
  - Сумісність з безіменним `layout=horizontal` (без регресії)
  - Іменовані групи всередині Category та Group
- [x] Лінт: 0 issues

**Файли:** `schema/uischema.go`, `parser/struct_parser.go`

**Результат:** Гнучке горизонтальне групування полів без необхідності створення вкладених структур.

---

## Етап 13 — Detail масиву (slice структур) ✅

Автоматична генерація UI Schema для елементів масиву структур (`[]Struct` / `[]*Struct`) через `options.detail`.

- [x] Визначення `[]struct` / `[]*struct` полів у `buildUIElements`
- [x] Генерація `options.detail` з `VerticalLayout` + Controls для полів елемента масиву
- [x] Scope в detail відносний: `#/properties/<field>`
- [x] Всі існуючі фічі працюють всередині detail: labels, readonly, multiline, rules, horizontal layout, вкладені структури (Group)
- [x] Примітивні slices (`[]string`, `[]int`) залишаються без detail
- [x] Порожні структури не створюють detail
- [x] Unit-тести:
  - `[]struct` та `[]*struct` → Control з `options.detail`
  - Примітивний slice → без detail
  - Detail з Category
  - HorizontalLayout всередині detail
  - Вкладена структура в елементі масиву
  - JSON серіалізація
  - Порожня структура → без detail
- [x] Лінт: 0 issues

**Файли:** `parser/struct_parser.go`

**Результат:** Масиви структур автоматично отримують UI Schema для своїх елементів через `options.detail`.

---

## Зведена таблиця

| Етап | Назва                          | Пріоритет | Залежність |
|------|--------------------------------|-----------|------------|
| 0    | Ініціалізація проєкту          | 🔴 High   | —          |
| 1    | JSON Schema з Go-структур      | 🔴 High   | Етап 0     |
| 2    | Підтримка struct-тегів         | 🔴 High   | Етап 1     |
| 3    | Генерація UI Schema            | 🔴 High   | Етап 1     |
| 4    | Rules (умовна логіка)          | 🟡 Medium | Етап 3     |
| 5    | Генерація з JSON               | 🟡 Medium | Етап 1     |
| 6    | HTTP API                       | 🟡 Medium | Етап 1-5   |
| 7    | Продуктивність та якість       | 🟡 Medium | Етап 1-6   |
| 8    | Розширення                     | 🟢 Low    | Етап 7     |
| 9    | Validation Constraints         | 🔴 High   | Етап 2     |
| 10   | HorizontalLayout               | 🔴 High   | Етап 3     |
| 11   | Rules / i18n на Layout         | 🟡 Medium | Етап 4, 10 |
| 12   | Іменовані Layout-групи ✅       | 🟡 Medium | Етап 10    |
| 13   | Detail масиву ✅                 | 🔴 High   | Етап 3     |
