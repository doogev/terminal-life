# Terminal Life — Roadmap

Milestones are ordered so that **something is on screen from day one**, and each one has an acceptance criterion that is a command plus an observable result — not a vibe.

---

## M0 — A world that grows

**Build**: `world` package, plant growth, ANSI renderer, frame loop, flags (`--w --h --seed --fps`).

**Acceptance**:

```bash
go run ./cmd/life --seed 1
```

- The grid starts as scattered seed patches and fills to saturation within roughly 60 ticks.
- `Ctrl-C` exits without leaving the terminal in a broken state (cursor restored, alt-screen exited).
- Piping to a file produces **no** escape codes: `go run ./cmd/life | head -c 200 | grep -c $'\x1b'` → `0`.

**Measured on first run**: 92 k ticks/s headless at 120×60; piped output contains 0 ESC bytes; 8 unit tests green.

**Correction to an earlier claim**: growth does *not* spread outward as fronts. Sunlight is applied uniformly to every cell (DESIGN §3), so the whole grid brightens in parallel and the seed patches are only distinguishable for the first few ticks via the render visibility threshold. A true spatial front would require neighbourhood-dependent growth, which changes the energy model — recorded as an open question rather than done at M0.

**Also expected**: with no consumers, the field saturates in ~60 ticks and then sits completely still. M0 has nothing that eats. This is not a stall.

---

## M1 — Herbivores and the first boom-bust

**Build**: `Agent`, `Genome`, energy accounting, sensing, movement, feeding, asexual reproduction, death. Spawn 10 herbivores.

**Acceptance**:

```bash
go run ./cmd/life --seed 1 --metrics out.jsonl --ticks 20000
```

- Population rises from 10 to a plateau, then oscillates. Any of: stable band, damped, or crash-and-recover — all acceptable, all recorded.
- `out.jsonl` shows `sun_in ≈ metabolism_out` within 10 % on the mean (the leak invariant from DESIGN §6).
- Same seed twice → identical `sha256sum`.

**Decision point**: if herbivores always strip the world flat and die within 2 k ticks regardless of seed, `Bite` and `SunRate` are mismatched — fix parameters before adding predators, otherwise M2 results are uninterpretable.

---

## M2 — Predators and the food chain

**Build**: predator kind, prey transfer, corpse decay, per-species gene means in metrics.

**Acceptance**:

- Predator count is non-zero for at least 5 k ticks on ≥ 1 of 5 seeds.
- Metrics show **out-of-phase** herbivore and predator peaks on at least one seed (the classic delayed oscillation). If it never appears, document which parameter was suspected, and what was tried — a negative result written down is still a result.
- No NaN, no negative energy, no agents outside bounds across 100 k ticks (`go test ./...` invariant suite).

---

## M3 — Headless and observable

**Build**: `--headless`, JSONL flush discipline, terminal sparkline for the three populations, `--params` flag overrides.

**Acceptance**:

```bash
go run ./cmd/life --headless --ticks 200000 --seed 3 --metrics run3.jsonl
wc -l run3.jsonl          # ≈ ticks / MetricsEvery
```

- Runs 200 k ticks in under a minute on the target machine.
- Sparkline renders three rows without corrupting the frame.
- Every flag in DESIGN §6 is settable; defaults reproduce the documented behaviour.

---

## M4 — Polish (optional)

- Palette tuning; plant gradient legible on both dark and light terminals.
- Parameter presets: `--preset boom-bust`, `--preset predator-crash`.
- Replay from seed at a chosen tick for inspection (`--until N`).

Explicitly **not** planned: save/load world state, web UI, multi-species, learning agents. Those belong to a different project.

---

## Risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Predators always collapse | M2 acceptance fails | Raise predator `BaseCost`, lower `TransferEfficiency`; if still failing, record it as a finding and move on |
| Sensing cost at scale | Frame budget blown | Bucketed spatial hash (see ARCHITECTURE §6) |
| Determinism silently broken by map iteration | Metrics never match | Keep maps out of the state path; the hash test in CI catches it |
| Gene drift unreadable | Metrics useless | Only 3 genes; per-species means; add entropy measure if needed (DESIGN §11) |

## Definition of done for v1

`go test ./...` green, M0–M3 acceptance commands all pass, and a 200 k-tick run produces a metrics file whose plant/predator/herbivore curves are worth looking at twice.
