# Contributing to gsap

Thanks for your interest in contributing! This project is a Go library for parsing messy LLM JSON into typed structs, and contributions of all kinds are welcome.

## Getting Started

1. Fork the repository and clone your fork.
2. Create a feature branch from `main`:
   ```
   git checkout -b my-feature
   ```
3. Install dependencies:
   ```
   go mod tidy
   ```

## Making Changes

- Follow existing code patterns and conventions.
- Run `gofmt` on all changed files (or use `goimports`).
- Run `go vet ./...` and fix any warnings.
- Keep commits focused and write clear commit messages.

## Testing

Run the full test suite before submitting:

```
go test -v -race ./...
```

Requirements:

- All existing tests must pass.
- New features and bug fixes should include tests.
- Test edge cases, especially around malformed JSON and type coercion.

## Submitting a Pull Request

1. Push your branch to your fork.
2. Open a pull request against `main`.
3. Describe what your change does and why.
4. Link any related issues.

## Code Style

- Use `gofmt` for formatting.
- Use `go vet` for static analysis.
- Exported functions and types need doc comments.
- Prefer clear variable names over short abbreviations.
- Handle errors explicitly; do not discard them.

## Questions?

Open an issue if you have questions or want to discuss a larger change before starting work.
