package world

// Kind distinguishes the trophic roles. Omnivore is reserved so that adding
// it later does not require rewriting switch statements everywhere.
type Kind uint8

const (
	Herbivore Kind = iota
	Predator
	Omnivore // not simulated in v1
)

func (k Kind) String() string {
	switch k {
	case Herbivore:
		return "herbivore"
	case Predator:
		return "predator"
	case Omnivore:
		return "omnivore"
	}
	return "unknown"
}

// Genome is deliberately three genes. More genes make the drift curves
// unreadable on a sparkline; see DESIGN.md §4.
type Genome struct {
	Speed  int8  // cells per tick, 1..3
	Vision int8  // search radius, 1..12
	Repro  int16 // energy at which the agent buds, 40..160
}

// Gene bounds. Mutation clamps to these.
const (
	MinSpeed, MaxSpeed   = 1, 3
	MinVision, MaxVision = 1, 12
	MinRepro, MaxRepro   = 40, 160
)

// Agent is one creature. There is no behaviour field: what an agent "does"
// emerges from genome plus energy budget.
type Agent struct {
	X, Y   int
	Energy float32
	Age    int
	G      Genome
	Kind   Kind
}

// RandomGenome draws a starting genome in the documented init ranges.
func RandomGenome(w *World) Genome {
	return Genome{
		Speed:  int8(1 + w.RNG.IntN(2)),    // 1..2
		Vision: int8(3 + w.RNG.IntN(4)),    // 3..6
		Repro:  int16(60 + w.RNG.IntN(41)), // 60..100
	}
}

// Mutate returns a copy of g with one randomly chosen gene perturbed by -1..+1
// and clamped. Mutation rate is fixed, not a gene: a self-tuning mutation gene
// collapses every run into a mutator attractor.
func Mutate(w *World, g Genome) Genome {
	switch w.RNG.IntN(3) {
	case 0:
		g.Speed = clampInt8(g.Speed+int8(w.RNG.IntN(3)-1), MinSpeed, MaxSpeed)
	case 1:
		g.Vision = clampInt8(g.Vision+int8(w.RNG.IntN(3)-1), MinVision, MaxVision)
	default:
		g.Repro = clampInt16(g.Repro+int16(w.RNG.IntN(3)-1), MinRepro, MaxRepro)
	}
	return g
}

func clampInt8(v, lo, hi int8) int8 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampInt16(v, lo, hi int16) int16 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
