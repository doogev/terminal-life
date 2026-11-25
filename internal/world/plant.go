package world

// Sunlight and plant growth. Every joule an agent will ever eat traces back
// to the line in GrowPlants; if that invariant is broken, the simulation is
// running on free energy and the metrics mean nothing.

// GrowPlants adds SunRate to every cell, saturating at MaxPlantEnergy.
// The saturating form (rate proportional to the gap) keeps cells from
// exploding and gives growth fronts their characteristic rounded shape.
func (w *World) GrowPlants() {
	max := w.P.MaxPlantEnergy
	if max <= 0 {
		return
	}
	rate := w.P.SunRate
	for i, v := range w.Plants {
		grown := v + rate*(1-v/max)
		if grown > max {
			grown = max
		}
		w.Plants[i] = grown
	}
}

// SeedPlants scatters n initial plant patches so the first frame is not empty.
// Each seed sets one cell to half saturation; growth does the rest.
func SeedPlants(w *World, n int) {
	if n <= 0 {
		return
	}
	half := w.P.MaxPlantEnergy * 0.5
	for i := 0; i < n; i++ {
		x := w.RNG.IntN(w.P.W)
		y := w.RNG.IntN(w.P.H)
		w.Plants[w.Index(x, y)] = half
	}
}

// Graze removes up to Bite energy from a cell and returns what was taken.
// Energy is conserved: what leaves the cell is what the caller gains.
func (w *World) Graze(x, y int) float32 {
	if !w.InBounds(x, y) {
		return 0
	}
	i := w.Index(x, y)
	avail := w.Plants[i]
	if avail <= 0 {
		return 0
	}
	take := min(avail, w.P.Bite)
	w.Plants[i] = avail - take
	return take
}

// DropCorpse returns a fraction of a dead agent's body energy to the cell as
// plant matter. Without this, every death is a pure sink and a predator crash
// becomes irreversible. See DESIGN.md §5.
func (w *World) DropCorpse(x, y int, energy float32) {
	if !w.InBounds(x, y) || energy <= 0 {
		return
	}
	const corpseFraction = 0.3
	i := w.Index(x, y)
	w.Plants[i] = min(w.Plants[i]+energy*corpseFraction, w.P.MaxPlantEnergy)
}
