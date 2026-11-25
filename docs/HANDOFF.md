# Handoff — read this first in a fresh session

You are picking up **terminal-life**, a terminal ecosystem simulator. This file exists so that a new session (or a future you) can resume without re-deriving decisions from scratch.

## Read in this order

0. `STATUS.md` — where the project actually is right now: milestone, measured numbers, known gaps.
1. `README.md` — what it is, in 30 seconds.
2. `docs/DESIGN.md` — world model, genome, tick pipeline, energy accounting, parameters. **This is the source of truth for behaviour.**
3. `docs/ARCHITECTURE.md` — package layout, dependency rule, performance budget.
4. `docs/ROADMAP.md` — current milestone and its acceptance command.
5. This file — conventions and current state.

## Current state

- **Design phase complete.** No Go code has been written yet.
- Repository: `github.com/doogev/terminal-life` (private).
- **No Go toolchain on this machine yet** (`go: command not found`). Install Go before M0 — do not start writing code you cannot compile.
- Next action: **M0** — world grid, plant growth, ANSI renderer, frame loop.

## Working conventions

- **All repository content in English**: code, comments, commit messages, docs, issues, PR titles.
- Repo lives at `~/git/doogev/terminal-life`. Identity is injected automatically by `includeIf`; do not override it.
- Use `gh-doogev` for every GitHub operation. Bare `gh` is intentionally unauthenticated on this machine.
- Standard library only. No `go get` unless there is a written justification in the commit message.
- Commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:` prefixes; imperative mood.

## Non-negotiables

1. **Determinism.** No map iteration in the state path, no goroutines touching the world, one seeded RNG passed explicitly. The 10 k-tick hash test must stay green.
2. **Dependency direction.** `render` and `metrics` are read-only consumers of `world`. Only `sim` mutates.
3. **`sim` must stay headless-testable.** Any terminal I/O belongs in `cmd/life`.
4. **Document negative results.** If a milestone's acceptance criterion fails (e.g. no predator oscillation appears), write down what was tried and what was observed, in `docs/ROADMAP.md` next to that milestone.

## When you finish a milestone

1. Run its acceptance command for real; paste the observed output into the commit message.
2. Update the status line at the top of `docs/ROADMAP.md`.
3. Commit and push with `gh-doogev`-authenticated git (plain `git push` works, the remote is SSH).
