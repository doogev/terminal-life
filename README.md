# Terminal Life

A character-based ecosystem that runs in your terminal. Sunlight grows plants, herbivores eat plants, predators eat herbivores. Every creature carries a genome. Nothing about the food chain is scripted — the population dynamics are whatever the rules happen to produce.

```
. . , . . .        ,   , .          , = plant
. , o . . ,    .       o .          o = herbivore
, . . . O , .      o       .        O = predator
. , . o . . ,  ,       O   .
```

## Status

**M0 complete** — the plant field grows and renders. Herbivores arrive in M1.

[STATUS.md](STATUS.md) holds the measured numbers: test count, throughput, known gaps, and what the next milestone needs.

## Quick start

```bash
go run ./cmd/life                 # watch it in your terminal
go run ./cmd/life --seed 42       # same seed, same run, every time
go run ./cmd/life --headless --ticks 100000
go run ./cmd/life --w 60 --h 20 --plants 120 --sun 0.01
```

Requires Go 1.27+ (`~/apps/go/bin` on the development machine). No third-party dependencies.

## Design principles

1. **Rules, not behaviour.** No agent has a "flee" routine. Genes plus energy budgets produce whatever strategy survives.
2. **Determinism is a feature.** Same seed and same tick count must produce byte-identical metrics. Otherwise the experiment is not repeatable and the history is worthless.
3. **No dependencies.** Standard library only. One binary, no service, no database.
4. **Observe, don't tune.** The point is to find out what happens, not to force a stable equilibrium.

## License

MIT
