---
name: refactor
description: Refactor code to improve readability, reduce duplication, and follow best practices.
allow_implicit_invocation: true
---

Refactor the specified code. Follow these principles:

1. **Preserve behavior** — refactoring must not change external behavior
2. **Single responsibility** — each function/method does one thing
3. **DRY** — eliminate duplicated logic by extracting shared helpers
4. **Naming** — use clear, descriptive names that convey intent
5. **Simplify** — reduce nesting, flatten conditionals, remove dead code
6. **Verify** — run existing tests after refactoring to confirm no regressions
