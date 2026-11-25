# Terminal Life

A character-based ecosystem that runs in your terminal. Sunlight grows plants, herbivores eat plants, predators eat herbivores. Every creature carries a genome. Nothing about the food chain is scripted — the population dynamics are whatever the rules happen to produce.

```
. . , . . .        ,   , .          , = plant
. , o . . ,    .       o .          o = herbivore
, . . . O , .      o       .        O = predator
. , . o . . ,  ,       O   .
```

## Status

**Design phase.** No code yet. The build order and acceptance criteria live in [docs/ROADMAP.md](docs/ROADMAP.md).

## Documentation

| Doc | Contents |
| --- | --- |
| [docs/DESIGN.md](docs/DESIGN.md) | World model, genome, tick pipeline, energy accounting, parameters |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Module layout, dependency rules, data flow, performance budget |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Milestones M0–M4, acceptance criteria, known failure modes |

## Quick start (once M1 lands)

```bash
go run ./cmd/life                 # watch it live in your terminal
go run ./cmd/life --seed 42       # same seed, same run, every time
go run ./cmd/life --headless --ticks 100000 --metrics out.jsonl
```

## Design principles

1. **Rules, not behaviour.** No agent has a "flee" routine. Genes plus energy budgets produce whatever strategy survives.
2. **Determinism is a feature.** Same seed and same tick count must produce byte-identical metrics. Otherwise the experiment is not repeatable and the history is worthless.
3. **No dependencies.** Standard library only. One binary, no service, no database.
4. **Observe, don't tune.** The point is to find out what happens, not to force a stable equilibrium.

## License

MIT
