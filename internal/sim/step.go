// Package sim owns the tick pipeline. It is the only package allowed to
// mutate agent state, and it never performs I/O, so it stays headless-testable.
package sim

import "github.com/doogev/terminal-life/internal/world"

// Step advances the world by exactly one tick.
//
// The stage order is fixed and load-bearing (DESIGN.md §5). M0 implements
// stages 1 and the tick counter; later milestones insert the remaining
// stages between them without reordering.
func Step(w *world.World) {
	// 1. sunlight
	w.GrowPlants()

	// 2..7: sense, move, feed, metabolise, reproduce, die (M1/M2)

	w.Tick++
}
