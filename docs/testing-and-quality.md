# Testing and Code Quality Standards

This document outlines the testing and code quality standards for the DJ project.

## Test Coverage

### Minimum Coverage Thresholds

- **Overall project coverage**: 30% minimum
- **New packages**: Aim for 50%+ coverage
- **Critical packages** (permissions, hooks, checkpoint): 70%+ coverage

### Coverage Reporting

Coverage is automatically calculated and reported in CI:

```bash
# Run tests with coverage locally
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage report
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out
```

Coverage reports are uploaded to Codecov on every CI run.

## Code Complexity Limits

We enforce complexity limits to maintain code readability and maintainability:

### Cyclomatic Complexity

- **Maximum**: 25 per function (excluding Update methods)
- **Tool**: `gocyclo`
- **Exclusions**:
  - Test files (`*_test.go`)
  - `Update` methods (Bubble Tea event handlers)
  - `Dispatch` methods (orchestration logic)

**Check locally:**
```bash
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
gocyclo -over 25 .
```

### Cognitive Complexity

- **Maximum**: 40 per function (excluding Update/Dispatch/async handlers)
- **Tool**: `gocognit`
- **Exclusions**:
  - Test files (`*_test.go`)
  - `Update` methods
  - `Dispatch` methods
  - Async handlers (functions with `wait` in name)

**Check locally:**
```bash
go install github.com/uudashr/gocognit/cmd/gocognit@latest
gocognit -over 40 .
```

## Testing Best Practices

### Test Organization

- **File naming**: `*_test.go` in the same package
- **Table-driven tests**: Use subtests with `t.Run()` for multiple scenarios
- **Test naming**: `Test<FunctionName>_<Scenario>`

Example:
```go
func TestLoadFromFile_InvalidJSON(t *testing.T) {
    // Test implementation
}
```

### What to Test

1. **Happy paths**: Normal operation with valid inputs
2. **Edge cases**: Boundary conditions, empty inputs, nil values
3. **Error handling**: Invalid inputs, I/O errors, network failures
4. **Concurrency**: Race conditions, mutex protection (where applicable)

### Test Structure

Follow the AAA pattern:

```go
func TestExample(t *testing.T) {
    // Arrange
    input := setupTestData()

    // Act
    result := functionUnderTest(input)

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Mocking and Test Doubles

- Use interfaces to enable testing
- Create test doubles in `*_test.go` files
- Use `t.TempDir()` for filesystem tests
- Use `httptest` package for HTTP testing

## Continuous Integration

Our CI pipeline runs on every pull request and includes:

1. **Build**: Verify code compiles (`go build ./...`)
2. **Tests**: Run all tests with race detection (`go test -race`)
3. **Coverage**: Calculate and enforce minimum thresholds
4. **Complexity**: Check cyclomatic and cognitive complexity
5. **Vet**: Run static analysis (`go vet ./...`)
6. **Codecov**: Upload coverage reports

### CI Workflow

Located at `.github/workflows/ci.yml`, the workflow:

- Runs on Ubuntu latest
- Uses Go version from `go.mod`
- Installs complexity tools (`gocyclo`, `gocognit`)
- Generates coverage profiles
- Reports metrics and warnings

## Writing New Tests

When adding new features:

1. **Write tests first** (TDD approach recommended)
2. **Cover happy path** and at least one error case
3. **Run tests locally** before committing:
   ```bash
   go test ./... -v -race
   ```
4. **Check coverage** of your changes:
   ```bash
   go test -coverprofile=coverage.out ./path/to/package
   go tool cover -func=coverage.out
   ```

## Packages Needing More Tests

Current coverage by package:

| Package | Coverage | Priority |
|---------|----------|----------|
| `hooks` | 93.3% | ✅ Good |
| `checkpoint` | 77.1% | ✅ Good |
| `modes` | 76.7% | ✅ Good |
| `memory` | 71.4% | ✅ Good |
| `tui/theme` | 50.0% | ⚠️ Could improve |
| `mentions` | 45.3% | ⚠️ Could improve |
| `skills` | 32.0% | 🔴 Needs work |
| `api` | 23.1% | 🔴 Needs work |
| `tui/components` | 11.5% | 🔴 Needs work |
| `agents` | 11.4% | 🔴 Needs work |
| `lsp` | 6.0% | 🔴 Needs work |
| `config` | 5.9% | 🔴 Needs work |
| `tui`, `tui/screens`, `mcp` | 0.0% | 🔴 Critical |

## Quality Goals

### Short-term (Next Release)

- ✅ Establish minimum coverage threshold (30%)
- ✅ Add complexity checks to CI
- ✅ Write tests for `theme`, `api/tracker`, `skills/matcher`
- 🎯 Bring all packages to >10% coverage

### Medium-term

- 🎯 Overall coverage >40%
- 🎯 All packages >20% coverage
- 🎯 Critical paths (permissions, execution) >80%

### Long-term

- 🎯 Overall coverage >50%
- 🎯 Integration tests for TUI flows
- 🎯 Benchmark tests for performance-critical paths

## Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table-driven tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Effective Go - Testing](https://go.dev/doc/effective_go#testing)
- [Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)
