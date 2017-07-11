package solve_test

import (
	"fmt"
	"github.com/bertbaron/solve"
)

type state struct {
	vector [5]byte
	cost   int
}

func (s state) Expand(ctx solve.Context) []solve.State {
	var steps []solve.State
	for i := 0; i < len(s.vector)-1; i++ {
		copy := s.vector
		copy[i], copy[i+1] = copy[i+1], copy[i]
		steps = append(steps, state{copy, s.cost + 1})
	}
	return steps
}

func (s state) IsGoal(ctx solve.Context) bool {
	for i := 1; i < len(s.vector); i++ {
		if s.vector[i-1] > s.vector[i] {
			return false
		}
	}
	return true
}

func (s state) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s state) Heuristic(ctx solve.Context) float64 {
	return 0
}

func sameState(a, b solve.State) bool {
	return a.(state).vector == b.(state).vector
}

// Finds the minimum number of swaps of neighbouring elements required to
// sort a vector
func Example() {
	s := state{[...]byte{3, 2, 5, 4, 1}, 0}
	result := solve.NewSolver(s).
		Algorithm(solve.IDAstar).
		Constraint(solve.NoLoopConstraint(10, sameState)).
		Solve()
	for _, st := range result.Solution {
		fmt.Printf("%v\n", st.(state).vector)
	}
	// Output:
	// [3 2 5 4 1]
	// [3 2 5 1 4]
	// [3 2 1 5 4]
	// [3 2 1 4 5]
	// [3 1 2 4 5]
	// [1 3 2 4 5]
	// [1 2 3 4 5]
}
