package main

import (
	"fmt"
)

type swapState struct {
	vector []byte
	cost   float64
	op     int
}

// returns a copy of the given vector, where the element at index is swapped with its right neighbour
func swap(vector []byte, index int) []byte {
	cp := make([]byte, len(vector), len(vector))
	copy(cp, vector)
	cp[index], cp[index + 1] = cp[index + 1], cp[index]
	return cp
}

func (v swapState) Expand() []State {
	n := len(v.vector) - 1
	steps := make([]State, n, n)
	for i := 0; i < n; i++ {
		steps[i] = swapState{swap(v.vector, i), v.cost + 1.0, i}
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

func printSolution(node *Node) {
	if node == nil {
		return
	}
	printSolution(node.parent)
	fmt.Println(node.state)
}

func main() {
	state := swapState{[]byte{1, 4, 5, 2, 3}, 0.0, -1}
	solution := Solve(state, IDAstar, 20.0)
	fmt.Printf("visited: %d, expanded %d\n", solution.Visited, solution.Expanded)
	if solution.Solution == nil {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found in %.0f steps\n", solution.Solution.state.Cost())
		printSolution(solution.Solution)
	}
}