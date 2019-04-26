package geom

import (
	"fmt"
	"math"
)

// ContextualObstacle captures how much of an obstacle this is to the navigator.
// A bird can fly right over a tree, a snake is not impeded by a swamp. A horse
// runs fastest when the ground is level and clear. The Cost multiplies the
// normal traversal time. A Cost of 2 implies that taking this path is twice as
// long as it normally would be. A cost of Infinity marks something that is completely impassable.
type ContextualObstacle struct {
	M, N int

	Cost float64
}

func reconstruct(prevs map[*Hex]*Hex, current *Hex) ([]*Hex, error) {
	result := []*Hex{current}
	n, ok := prevs[current]
	for ok {
		result = append(result, n)

		// Next!
		n, ok = prevs[n]
	}
	for i := len(result)/2 - 1; i >= 0; i-- {
		opp := len(result) - 1 - i
		result[i], result[opp] = result[opp], result[i]
	}
	return result, nil
}

// heuristic determines the comparitive "as the crow flies" distance between two
// Hexes, ignoring obstacles.
func heuristic(a, b *Hex) float64 {
	// pythagorean theorum, minus the sqrt.
	return math.Pow(a.X()-b.X(), 2) + math.Pow(a.Y()-b.Y(), 2)
}

// Navigate a path from start to the goal, avoiding Impassable Hexes.
func Navigate(start, goal *Hex, obstacles []ContextualObstacle) ([]*Hex, error) {
	oneStep := heuristic(&Hex{M: 0, N: 0}, &Hex{M: 0, N: 1})

	closed := map[Key]interface{}{}
	open := map[*Hex]interface{}{
		start: struct{}{},
	}
	cameFrom := map[*Hex]*Hex{}
	costs := map[*Hex]float64{
		start: 0,
	}
	guesses := map[*Hex]float64{
		start: heuristic(start, goal),
	}

	for len(open) > 0 {
		var current *Hex
		low := math.MaxFloat64
		for k := range open {
			if guesses[k] < low {
				current = k
				low = guesses[k]
			}
		}
		if current == goal {
			return reconstruct(cameFrom, current)
		}

		if current == nil {
			break
		}

		delete(open, current)
		closed[Key{M: current.M, N: current.N}] = struct{}{}

		for _, n := range current.Neighbors() {
			if _, ok := closed[Key{M: n.M, N: n.N}]; ok {
				continue
			}

			tentative := costs[current] + oneStep

			// The cost of passing through this hex might be affected by any
			// obstacles occupying the Hex.
			for _, o := range obstacles {
				if o.M == n.M && o.N == n.N {
					if o.Cost == math.Inf(0) {
						tentative = math.MaxFloat64
					} else {
						tentative *= o.Cost
					}
					break
				}
			}

			if _, ok := open[n]; !ok {
				open[n] = struct{}{}
			} else if tentative >= costs[n] {
				continue
			}

			cameFrom[n] = current
			costs[n] = tentative
			guesses[n] = costs[n] + heuristic(n, goal)
		}
	}
	return nil, fmt.Errorf("no path available from %d,%d to %d,%d", start.M, start.N, goal.M, goal.N)
}

/*
This Navigation does not support large characters that occupy more than one
hex at a time.

I think the interface should go from

	func Navigate(start, goal *Hex, obstacles []ContextualObstacle) ([]*Hex, error) {

to

	func Navigate(start []*Hex, m, n int, obstacles []ContextualObstacle) ([]*Hex, error) {

where start now accepts a slice of hexes that the character occupies, and m,n
represent the goal by the number of hexes to offset the each starting hex by.

I wonder what this means for detecting an M,N offset in terms of translating
mouse coordinates?

Potential issues:

- We need to check whether any obstacle is only blocked by the character we
are pathfinding *for*, and would not be an obstacle if the character was
moving.

Another option to explore would be to codify small, medium and large sized
units, and have separate coordinate systems for each. This might be easier to
implement side-by-side with the existing logic, i.e:

	func Navigate(start, goal *Hex4, obstacles []ContextualObstacle) ([]*Hex4, error) {
	func Navigate(start, goal *Hex7, obstacles []ContextualObstacle) ([]*Hex7, error) {

Where Hex4 is something like

type Hex4 struct {
	O,P int
	[]*Hex hexes
	[]*Hex4 neighbors
}

and Hex7 looks like

type Hex7 struct {
	Q,R int
	[]*Hex hexes
	[]*Hex7 neighbors
}
*/