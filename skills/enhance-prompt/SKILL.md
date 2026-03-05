---
name: enhance-prompt
description: Restructures a raw prompt into a detailed Codex task with acceptance criteria, constraints, scope, and reasoning hints.
allow_implicit_invocation: false
---

Transform the user's prompt into a structured Codex task. Include:

1. **Role context** — who Codex is acting as for this task
2. **Acceptance criteria** — explicit numbered list of what "done" looks like
3. **File/surface scope** — which files or modules are in/out of scope
4. **Constraints** — no silent error catches, DRY principle, propagate errors explicitly
5. **Cadence** — acknowledge task, plan in 1-2 sentences, then execute
6. **Reasoning effort hint** — suggest "low" / "medium" / "high" based on complexity
7. **Edge cases** — enumerate 2-3 foreseeable failure modes
