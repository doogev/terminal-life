// Package world holds the simulation state: the plant field and the agents.
// It deliberately knows nothing about ticks, rendering or output.
package world

import "math/rand/v2"

// Params are the tunable constants of a run. Defaults live in DefaultParams.
// Passing Params by value into New makes it trivial to add a config file later.
type Params struct {
	W, H               int
	SunRate            float32 // energy added per cell per tick
	MaxPlantEnergy     float32 // per-cell saturation
	BaseCost           float32 // metabolic cost per tick, all agents
	SpeedCost          float32 // extra cost per spd^2
	Bite               float32 // max plant energy eaten per feeding
	TransferEfficiency float32 // fraction of prey energy gained by predator
	ReproCost          float32 // fraction of energy paid to offspring
	MaxAge             int
	MetricsEvery       int
}

// DefaultParams returns the documented defaults from DESIGN.md §6.
func DefaultParams() Params {
	return Params{
		W:                  120,
		H:                  60,
		SunRate:            0.02,
		MaxPlantEnergy:     1.0,
		BaseCost:           0.02,
		SpeedCost:          0.02,
		Bite:               0.25,
		TransferEfficiency: 0.8,
		ReproCost:          0.5,
		MaxAge:             2000,
		MetricsEvery:       100,
	}
}

// World is the whole simulation state. Agents are stored in a slice rather
// than in the grid so that iteration order is stable (determinism) and so
// that transient cell sharing during a tick is easy to resolve.
type World struct {
	P      Params
	Plants []float32
	Agents []Agent
	RNG    *rand.Rand
	Tick   int
}

// New builds an empty world: no plants, no agents.
func New(p Params, seed uint64) *World {
	return &World{
		P:      p,
		Plants: make([]float32, p.W*p.H),
		Agents: make([]Agent, 0, 256),
		RNG:    rand.New(rand.NewPCG(seed, 0x9E3779B97F4A7C15)),
	}
}

// Index maps grid coordinates to a slice index. Callers must check InBounds.
func (w *World) Index(x, y int) int { return y*w.P.W + x }

// InBounds reports whether (x, y) is inside the grid. Edges are walls.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < w.P.W && y < w.P.H
}

// PlantAt returns the plant energy in a cell, or 0 outside the grid.
func (w *World) PlantAt(x, y int) float32 {
	if !w.InBounds(x, y) {
		return 0
	}
	return w.Plants[w.Index(x, y)]
}

// TotalPlantEnergy is the reservoir size, used by metrics and the leak check.
func (w *World) TotalPlantEnergy() float64 {
	var sum float64
	for _, v := range w.Plants {
		sum += float64(v)
	}
	return sum
}
