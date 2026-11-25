package world

import (
	"math"
	"testing"
)

func testParams() Params {
	p := DefaultParams()
	p.W, p.H = 8, 4
	return p
}

func TestGrowPlantsSaturates(t *testing.T) {
	p := testParams()
	w := New(p, 1)
	for i := range w.Plants {
		w.Plants[i] = p.MaxPlantEnergy
	}
	w.GrowPlants()
	for i, v := range w.Plants {
		if v > p.MaxPlantEnergy+1e-6 {
			t.Fatalf("cell %d exceeded saturation: %v", i, v)
		}
	}
}

func TestGrowPlantsIncreasesEmptyCells(t *testing.T) {
	p := testParams()
	w := New(p, 1)
	w.GrowPlants()
	for i, v := range w.Plants {
		if v <= 0 {
			t.Fatalf("cell %d did not grow: %v", i, v)
		}
	}
}

func TestGrazeConservesEnergy(t *testing.T) {
	p := testParams()
	w := New(p, 1)
	w.Plants[w.Index(2, 2)] = p.MaxPlantEnergy

	before := w.TotalPlantEnergy()
	taken := w.Graze(2, 2)
	after := w.TotalPlantEnergy()

	if taken <= 0 {
		t.Fatalf("expected to take energy, got %v", taken)
	}
	if math.Abs((before-float64(taken))-after) > 1e-5 {
		t.Fatalf("energy not conserved: before=%v taken=%v after=%v", before, taken, after)
	}
}

func TestGrazeOutsideGridIsNoop(t *testing.T) {
	p := testParams()
	w := New(p, 1)
	if got := w.Graze(-1, 0); got != 0 {
		t.Fatalf("expected 0 outside grid, got %v", got)
	}
}

func TestSeedPlantsDeterministic(t *testing.T) {
	p := testParams()
	a := New(p, 42)
	b := New(p, 42)
	SeedPlants(a, 10)
	SeedPlants(b, 10)
	for i := range a.Plants {
		if a.Plants[i] != b.Plants[i] {
			t.Fatalf("same seed diverged at cell %d: %v vs %v", i, a.Plants[i], b.Plants[i])
		}
	}
}

func TestMutateStaysInBounds(t *testing.T) {
	p := testParams()
	w := New(p, 7)
	g := Genome{Speed: MaxSpeed, Vision: MaxVision, Repro: MaxRepro}
	for i := 0; i < 5000; i++ {
		g = Mutate(w, g)
		if g.Speed < MinSpeed || g.Speed > MaxSpeed {
			t.Fatalf("speed out of bounds: %d", g.Speed)
		}
		if g.Vision < MinVision || g.Vision > MaxVision {
			t.Fatalf("vision out of bounds: %d", g.Vision)
		}
		if g.Repro < MinRepro || g.Repro > MaxRepro {
			t.Fatalf("repro out of bounds: %d", g.Repro)
		}
	}
}

func TestDropCorpseSaturates(t *testing.T) {
	p := testParams()
	w := New(p, 1)
	w.DropCorpse(1, 1, 1000)
	if got := w.PlantAt(1, 1); got > p.MaxPlantEnergy+1e-6 {
		t.Fatalf("corpse energy exceeded saturation: %v", got)
	}
}
