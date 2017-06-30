package main

import (
	"fmt"
	"github.com/bertbaron/solve"
)

type swapState struct {
	vector []byte
	cost   float64
	op     int
}

func newSwapState(vector []byte, cost float64, op int) swapState {
	return swapState{vector, cost, op}
}

func swapProblem(initialState []byte) swapState {
	return newSwapState(initialState, 0.0, -1)
}

// returns a copy of the given vector, where the element at index is swapped with its right neighbour
func swap(vector []byte, index int) []byte {
	cp := make([]byte, len(vector), len(vector))
	copy(cp, vector)
	cp[index], cp[index + 1] = cp[index + 1], cp[index]
	return cp
}

func (v swapState) Id() interface{} {
	return fmt.Sprintf("%v", v.vector)
}

func (v swapState) Expand() []solve.State {
	n := len(v.vector) - 1
	steps := make([]solve.State, n, n)
	for i := 0; i < n; i++ {
		steps[i] = newSwapState(swap(v.vector, i), v.cost + 1.0, i)
	}
	return steps
}

func (v swapState) IsGoal() bool {
	n := len(v.vector) - 1
	for i := 0; i < n; i++ {
		if v.vector[i] > v.vector[i + 1] {
			return false
		}
	}
	return true
}

func (v swapState) Cost() float64 {
	return v.cost
}

func (v swapState) Heuristic() float64 {
	return 0.0
}

func printSolution(node solve.Node) {
	if !node.Exists() {
		return
	}
	printSolution(node.Parent())
	swapstate := node.State().(swapState)
	for i, e := range swapstate.vector {
		if i > 0 {
			if i == swapstate.op + 1 {
				fmt.Print("x")
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Print(e)
	}
	fmt.Println()
}

func main() {
	state := swapProblem([]byte{5, 7, 4, 3, 6, 2, 1})
	fmt.Printf("Sorting %v in minimal number of swaps of neighbouring elements\n", state)
	solution := solve.NewSolver(state).
		Algorithm(solve.IDAstar).
		Constraint(solve.CHEAPEST_PATH).
		Limit(20).
		Solve()

	fmt.Printf("visited: %d, expanded %d\n", solution.Visited, solution.Expanded)
	if !solution.Solution.Exists() {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found in %.0f steps\n", solution.Solution.State().Cost())
		printSolution(solution.Solution)
	}
}
