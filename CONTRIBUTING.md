# Contributing to ui-json-schema

Thank you for considering contributing! This document explains how to get started.

## Prerequisites

- **Go 1.24+**
- **golangci-lint v2** — [install guide](https://golangci-lint.run/welcome/install/)
- **GNU Make**

## Getting Started

```bash
# Clone the repository
git clone https://github.com/holdemlab/ui-json-schema.git
cd ui-json-schema

# Verify everything works
make lint
make test-cover
make build
```

## Development Workflow

### 1. Create a branch

```bash
git checkout -b feat/my-feature master
```

### 2. Make your changes

Follow the existing project structure:

| Directory        | Purpose                                  |
|------------------|------------------------------------------|
| `schema/`        | Types, tag parsing, options, i18n        |
| `parser/`        | Schema generation (structs, JSON, OpenAPI)|
| `api/`           | HTTP handler and type registry           |
| `cmd/server/`    | Server entry point                       |

### 3. Run checks locally

```bash
make fmt          # Format code
make lint         # Run golangci-lint (must be 0 issues)
make test-cover   # Run tests with coverage (must be ≥ 80%)
make bench        # Run benchmarks
```

### 4. Commit using Conventional Commits

The CI auto-tags releases based on commit messages. Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

| Prefix              | Version bump | Example                                   |
|----------------------|-------------|-------------------------------------------|
| `feat:`             | **minor**    | `feat: add XML schema support`            |
| `fix:`              | **patch**    | `fix: handle nil pointer in parser`       |
| `feat!:` / `fix!:`  | **major**    | `feat!: redesign Options API`             |
| `BREAKING CHANGE`   | **major**    | commit body contains `BREAKING CHANGE`    |
| `docs:`, `chore:`, `refactor:`, `test:` | **patch** | `docs: update README`  |

### 5. Open a Pull Request

- Target the `master` branch.
- Ensure CI passes (lint → test → build).
- Include a clear description of the changes.

## Code Standards

### Tests

- Every new feature must have unit tests.
- Test coverage must remain **≥ 80%** (current: ~92%).
- Test files are excluded from `gocyclo`, `dupl`, `gosec`, `goconst`, `errcheck`, and `govet` linters.

### Linting

The project uses **golangci-lint v2** with a strict configuration (see `.golangci.yml`). Key linters:

`errcheck` · `govet` · `staticcheck` · `revive` · `gosec` · `gocyclo` · `dupl` · `goconst` · `misspell` · `errorlint`

Run `make lint` before committing — **0 issues** is required.

### Style

- US English in comments (`initialize`, not `initialise`).
- Exported functions must have doc comments ending with a period.
- Cyclomatic complexity limit: **15** per function.
- No external dependencies — the library uses only the standard library.

## Available Make Targets

```
make build        Compile the application
make run          Build and run the server
make test         Run all tests
make test-cover   Run tests with coverage report
make lint         Run golangci-lint
make bench        Run benchmarks
make fmt          Format code (gofmt + goimports)
make clean        Remove build artefacts
```

## CI Pipeline

Every push to `master` and every PR triggers:

1. **Lint** — golangci-lint v2
2. **Test** — tests with race detector + coverage ≥ 80% check
3. **Build** — compilation check

On merge to `master`, after all checks pass:

4. **Auto Tag** — creates a semver tag based on commit messages and updates the `VERSION` file.

## Reporting Issues

- Use GitHub Issues.
- Include Go version (`go version`), OS, and a minimal reproduction case.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
