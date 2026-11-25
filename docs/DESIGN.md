# Terminal Life — Design Document

- Status: draft v0.1
- Date: 2026-11-25
- Scope: v1 (M0–M3). Anything else is listed under *Future work*.

---

## 1. Purpose

A small simulation whose interesting output is **emergent**. The author writes three rules (plants grow, herbivores eat plants, predators eat herbivores) plus a genome and an energy budget. The food chain, the oscillation period, the extinction events — none of that is authored.

The secondary purpose is observability: a run must be reproducible from a seed and must emit machine-readable metrics, so that "what happened" can be answered months later without re-watching a terminal.

## 2. Goals and non-goals

### Goals

- A visually legible terminal render at 30–60 FPS on a 4-core / 4 GB CPU-only machine.
- Deterministic simulation: `(seed, params, tick count)` → identical metrics file, byte for byte.
- A genome small enough that its drift can be plotted and understood (3 genes in v1).
- Headless mode for long runs with JSONL metrics.

### Non-goals (v1)

- No GUI, no web frontend, no network access.
- No neural networks or learning. Evolution via mutation and selection only.
- No save/load of the live world state. Only the seed and parameters are persisted; replay is by re-running.
- No multi-threading. Concurrency would break determinism for no benefit at this scale.
- No sexual reproduction. Offspring are clones with mutations (asexual budding).

## 3. World model

A finite `W × H` grid, no wraparound (edges are walls — keeps gradients interesting and prevents a torus from hiding local extinction behind a seam).

Default dimensions: `120 × 60` cells. Rationale: terminal cells are roughly 1:2 (width:height), so 2:1 grid dimensions render as a visually square world. 120×60 = 7,200 cells, trivial to iterate per tick.

Each cell holds:

```go
type Cell struct {
    Plant float32 // 0 .. MaxPlantEnergy; 0 means empty
}
```

Agents are stored separately in a slice (not in the grid) to keep iteration order stable and to allow several agents to share a cell briefly during collision resolution.

### Plant layer

Plants are a scalar field, not entities. Every tick each cell gains:

```
growth = SunRate * (1 - Plant/MaxPlantEnergy)
```

A saturating growth curve means a cell approaches `MaxPlantEnergy` asymptotically instead of exploding. `SunRate` is the world's total energy income — **every joule an agent ever eats traces back to this line**. This is what keeps the simulation from being a free-energy toy.

Optional v1.1: shade parameter, so heavily grazed cells grow back slower.

## 4. Genome

Three genes. Deliberately few: with more genes, the drift becomes 5-dimensional and unreadable on a sparkline.

| Gene | Symbol | Range | Default init | Effect |
| --- | --- | --- | --- | --- |
| Speed | `spd` | 1–3 cells/tick | 1–2 random | Movement step size; metabolic cost is superlinear in speed |
| Vision | `vis` | 1–12 cells | 3–6 random | Radius of the nearest-food search |
| Repro threshold | `rep` | 40–160 energy | 60–100 random | Energy level at which the agent buds |

Mutation on reproduction, applied independently per gene:

```
delta ~ Uniform(-1, +1) rounded toward zero, then clamp to gene range
```

Mutation rate is **not** a gene in v1. A gene that controls its own mutation rate creates a runaway "mutator" attractor that swamps everything else, and it makes the runs much harder to interpret. Fixed at `1 gene mutated per birth, chosen at random`.

## 5. Tick pipeline

Order is fixed and load-bearing. Changing it changes the outcome, so it is documented here rather than discovered later.

```
tick(t):
  1. sunlight      — add SunRate to every cell, saturating at MaxPlantEnergy
  2. sense         — each agent finds nearest food within `vis`, using Chebyshev distance
  3. move          — step toward target; if none, random walk; walls clamp
  4. feed          — if the agent's cell has food, consume min(Plant, bite) or eat a herbivore
  5. metabolise    — energy -= BaseCost + SpeedCost*spd^2 ; age++
  6. reproduce     — if energy >= rep: pay half, spawn child on an adjacent free cell
  7. die           — energy <= 0 or age > MaxAge → remove; drop a corpse-energy fraction to the cell
  8. sample        — every MetricsEvery ticks, append one JSONL record
```

### Why sense before move

Agents that move first and sense afterwards spend a tick blind on arrival, which makes early runs feel sluggish and masks the effect of the vision gene. Sensing first also makes the `vis` gene's benefit measurable immediately.

### Why corpses return energy

Without corpse energy, every death is a pure energy sink, and a predator crash becomes irreversible. Returning 30 % of the body energy to the cell as plant matter gives the system a decay path and produces visible "grass grows where things died" behaviour.

## 6. Energy accounting

Per tick, per agent:

```
cost = BaseCost + SpeedCost * spd²
```

Superlinear in speed is deliberate: it makes "just be fast" a losing strategy past a point, so the speed gene has an interior optimum instead of running to the maximum.

Default values (all tunable via flags):

| Parameter | Default | Note |
| --- | --- | --- |
| `SunRate` | 0.02 / cell / tick | Total world income ≈ 144 energy/tick at 120×60 |
| `MaxPlantEnergy` | 1.0 | A full cell feeds one herbivore for a while |
| `BaseCost` | 0.02 / tick | Bare survival cost |
| `SpeedCost` | 0.02 | `spd=3` costs 0.18 + 0.02 = 0.20/tick, 10× a `spd=1` agent |
| `Bite` | 0.25 | Max plant energy removed per feeding action |
| `TransferEfficiency` | 0.8 | Predator gains 80 % of prey body energy |
| `ReproCost` | 0.5 | Fraction of energy paid to the child |
| `MaxAge` | 2000 ticks | Hard cap so lineages cannot be immortal |
| `MetricsEvery` | 100 ticks | Sampling interval |

**Sanity invariant**: long-run mean plant stock × SunRate should be within ~10 % of total agent metabolism. If it is not, the run is leaking energy somewhere and the numbers are not trustworthy. There is a unit test for this.

## 7. Rendering

- Double-buffered `[]rune` + `[]uint8` (colour index) of size `W*H`.
- Full redraw is fine at 7,200 cells; the diff optimisation is only worth it if profiling says so.
- One ANSI cursor-home plus one row-serialised frame per tick. No per-cell cursor moves.
- ANSI 256-colour palette: plants dim green gradient by energy, herbivores yellow, predators red, intensity scaled by energy.
- Frame rate: default 20 FPS (sleep to pace), flag `--fps`.
- Non-TTY (piped output) automatically switches to headless, so `| tee` does not fill the log with escape codes.

## 8. Determinism

Rules that make runs reproducible:

1. One `math/rand/v2` PCG source seeded from `--seed`, passed explicitly. No global rand.
2. Agents iterate in slice order; removals are done by compaction after the tick, never mid-iteration.
3. No map iteration in any code path that affects state.
4. Float32 is used consistently; no mixing with float64 in the hot path.
5. Metrics are written with fixed formatting (`%.4f`) so files diff cleanly.

Test: run `--headless --ticks 10000 --seed 7` twice, `sha256sum` the two JSONL files, expect a match. This is in CI.

## 9. Metrics

One JSON object per sample (JSONL), one line per `MetricsEvery` ticks:

```json
{"tick":1000,"plants":3821.44,"herb":84,"pred":12,
 "herb_genes":{"spd":1.42,"vis":4.8,"rep":73.1},
 "pred_genes":{"spd":2.1,"vis":7.3,"rep":95.0},
 "births":31,"deaths":27,"sun_in":144.0,"metabolism_out":142.6}
```

Design notes:

- `plants` is total plant energy (a float), not a cell count — it is the real reservoir.
- Gene means are population means, unweighted. Weighting by energy was considered and rejected: it makes lineage turnover invisible.
- `sun_in` minus `metabolism_out` is the leak detector referenced in §6.

Terminal sparklines (`▁▂▃▄▅▆▇█`) for the three populations are a v1.1 nicety, not required for M3.

## 10. Expected failure modes

Documented up front so that observing them is not mistaken for a bug:

1. **Predator overkill.** Predators eat herbivores faster than herbivores reproduce, both die, world becomes grass. Expected in early parameter sets. Mitigation: raise predator `BaseCost`, lower `TransferEfficiency`.
2. **Herbivore overshoot.** No predators → herbivores strip the plant layer → mass starvation → crash → regrowth → repeat. This is a feature; it is the classic Lotka–Volterra delayed oscillation.
3. **Vision runaway.** If vision is free, `vis` drifts to max and the population becomes homogeneous. `SpeedCost`-style cost for vision is a candidate v1.1 change.
4. **Immortal lineages.** Without `MaxAge`, a lucky genome never dies and the gene mean freezes. Hence the hard age cap.

## 11. Open questions

- Should predators be able to starve into herbivory (omnivory) if prey is absent? Deferred until M2 data exists.
- Is a spatial hash needed? At 200 agents, brute force is 40 k distance checks/tick — fine. Only revisit above ~1,000 agents.
- Should the metrics file also record an entropy measure of the gene distribution, to distinguish "converged" from "drifting"?

## 12. Glossary

| Term | Meaning |
| --- | --- |
| tick | One atomic simulation step; all agents act once |
| bud | Asexual reproduction event |
| bite | Maximum energy removed from a plant cell in one feeding |
| leak | Any energy leaving the system other than through metabolism/death |
| run | One `(seed, params)` execution from tick 0 |
