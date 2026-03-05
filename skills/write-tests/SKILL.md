---
name: write-tests
description: Generate comprehensive test suites for code with edge cases, mocks, and assertions.
allow_implicit_invocation: true
---

Write tests for the specified code. Follow these guidelines:

1. **Test structure** — use the project's existing test framework and patterns
2. **Coverage** — happy path, error cases, edge cases, boundary values
3. **Naming** — descriptive test names that explain the scenario being tested
4. **Isolation** — mock external dependencies, test one behavior per test
5. **Assertions** — use specific assertions, not just "no error"
6. **Setup/Teardown** — use proper test fixtures, clean up after tests
