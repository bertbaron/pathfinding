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

type dummyState float64

func (v dummyState) Id() interface{} {
	return v
}

func (v dummyState) Expand() []solve.State {
	n := 5
	steps := make([]solve.State, n, n)
	for i := 0; i < n; i++ {
		steps[i] = dummyState(v + 1.0)
	}
	return steps
}

func (v dummyState) IsGoal() bool {
	return v >= 11
}

func (v dummyState) Cost() float64 {
	return float64(v)
}

func (v dummyState) Heuristic() float64 {
	return 0.0
}
func printSolution(node solve.Node) {
	if !node.Exists() {
		return
	}
	printSolution(node.Parent())
	fmt.Println(node.State())
}

func main() {
	/*
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	*/

	state := swapProblem([]byte{5, 4, 3, 6, 2, 1})
	//state := dummyState(0.0)
	solution := solve.NewSolver(state).Algorithm(solve.IDAstar).Constraint(solve.CHEAPEST_PATH).Limit(20).Solve()
	fmt.Printf("visited: %d, expanded %d\n", solution.Visited, solution.Expanded)
	if !solution.Solution.Exists() {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found in %.0f steps\n", solution.Solution.State().Cost())
		//printSolution(solution.Solution)
	}

	//f, err = os.Create("mem.prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.WriteHeapProfile(f)
	//f.Close()
}
