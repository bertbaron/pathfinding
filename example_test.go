package solve_test

import (
	"github.com/bertbaron/solve"
	"fmt"
)

type state struct {
	vector [5]byte
	cost   int
	index  int
}

func (s state) Id() interface{} {
	return s.vector
}

func (s state) Expand() []solve.State {
	n := len(s.vector) - 1
	steps := make([]solve.State, n, n)
	for i := 0; i < n; i++ {
		child := state{s.vector, s.cost + 1, i}
		child.vector[i], child.vector[i + 1] = child.vector[i + 1], child.vector[i]
		steps[i] = child
	}
	return steps
}

func (s state) IsGoal() bool {
	n := len(s.vector) - 1
	for i := 0; i < n; i++ {
		if s.vector[i] > s.vector[i + 1] {
			return false
		}
	}
	return true
}

func (s state) Cost() float64 {
	return float64(s.cost)
}

func (s state) Heuristic() float64 {
	return 0
}

// Finds the minumum number of swaps of neighbouring elements required to
// sort a vector
func Example() {
	s := state{[...]byte{3, 2, 5, 4, 1}, 0, -1}
	result := solve.NewSolver(s).
		Algorithm(solve.IDAstar).
		Constraint(solve.NO_LOOP).
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
