# Status

**Last updated:** 2026-11-25
**Current milestone:** M0 complete — plant field and renderer only, no agents.
**Latest commit:** `a4a1ebb` (feat: M0 — plant field, ANSI renderer, frame loop)

This file is the quick "where are we" answer. It records **measured** facts, not intentions. Milestone definitions live in [ROADMAP.md](ROADMAP.md); behaviour spec lives in [DESIGN.md](DESIGN.md).

---

## Snapshot

| Item | State | How it was checked |
| --- | --- | --- |
| Milestone | M0 ✅ (M1 not started) | — |
| Unit tests | 11 passing | `go test ./... -v` |
| `gofmt` | clean | `gofmt -l .` → empty |
| `go vet` | clean | `go vet ./...` → no output |
| Headless throughput | ~92,355 ticks/s @ 120×60, single core | `go run ./cmd/life --headless --ticks 1000` |
| Piped output escapes | 0 ESC bytes | `go run ./cmd/life --w 30 --h 8 --ticks 5 \| tr -cd '\033' \| wc -c` |
| Binary size | 2.6 MB | `go build -o /tmp/life ./cmd/life` |
| Go source size | 734 lines | `find . -name '*.go' \| xargs wc -l` |
| Third-party deps | none | `go.mod` has no `require` block |
| Repository | public | `gh-doogev repo view` |

## What actually works

```bash
go run ./cmd/life                              # interactive, 20 fps, 120x60
go run ./cmd/life --seed 42                    # deterministic replay
go run ./cmd/life --w 80 --h 30 --plants 200 --sun 0.01
go run ./cmd/life --headless --ticks 10000     # summary only, no frames
go run ./cmd/life --ticks 3                    # piped: plain text frames
```

- Plant field grows with a saturating curve from a fixed sunlight rate.
- Three output modes chosen automatically: interactive (TTY), plain (piped, zero escape bytes), headless (no frames).
- `Ctrl-C` / `SIGTERM` restores the cursor and leaves the alternate screen.
- Determinism primitives are in place: seeded PCG, stable slice iteration, no map in the state path.
- `Graze`, `DropCorpse`, `Genome`, `Mutate` exist and are unit-tested, but nothing calls them yet.

## Known gaps (not bugs)

| Gap | Why it is like this | Milestone |
| --- | --- | --- |
| Nothing eats the plants | No agents exist | M1 |
| Field saturates in ~60 ticks then freezes | Zero consumption; physics is correct, the world is just idle | M1 |
| No spatial growth fronts | Sunlight is applied uniformly to every cell, so the whole grid brightens in parallel | open question |
| No metrics output | JSONL writer not written | M3 |
| `--fps` only affects interactive mode | Piped mode runs uncapped and will flood the pipe | by design |
| No terminal-resize handling | Grid is fixed; a terminal narrower than `--w` will wrap and look broken | not planned |
| No release binaries | Only `go run` / local `go build` | M4 at the earliest |
| No streaming `agents=` values | `summarize` always reports 0 agents | M1 |

## Claims that were corrected

Recorded so that the same wrong statement is not repeated in a future summary:

1. "Growth spreads outward like mould colonies." **Wrong.** Growth is uniform per cell (DESIGN §3). There are no fronts. A visibility threshold (`plantVisible = 0.05`) is what makes seed patches readable for the first few ticks.
2. "8 unit tests." **Wrong in the M0 commit message.** There are **11**. The commit is already pushed; the number is corrected here rather than rewriting public history.

## Next up: M1

Checklist for the herbivore milestone, in the order the code should be built:

- [ ] Spawn 10 herbivores with random genomes at startup (`--agents`)
- [ ] Implement sense → move → feed → metabolise stages in `sim.Step`
- [ ] Asexual budding with `Mutate`, respecting `ReproCost`
- [ ] Death by starvation and `MaxAge`; corpse energy via `DropCorpse`
- [ ] Energy-balance test: mean `sun_in` vs `metabolism_out` within 10 %
- [ ] Determinism test: two 10 k-tick runs, identical metrics bytes
- [ ] Re-check throughput with ~200 agents; frame budget is 5 ms

Acceptance command and expected observations are in [ROADMAP.md](ROADMAP.md) M1.

## Environment

| Item | Value |
| --- | --- |
| Go | go1.27.4 linux/amd64, installed at `~/apps/go`, on `PATH` via `~/.bashrc` |
| Machine | 4 cores / 4.3 GB, no GPU (CPU-only by design) |
| Repo path | `~/git/doogev/terminal-life` |
| GitHub CLI | `gh-doogev` only; bare `gh` is intentionally unauthenticated |

## How to update this file

On every milestone completion: bump `Last updated`, update the Snapshot row values with freshly measured numbers, move finished items out of "Known gaps", and rewrite "Next up". Do **not** add aspirational items here — those belong in ROADMAP.
