package main

import (
	"fmt"
	"github.com/bertbaron/solve"
	"sort"
)

const max_size = 16

type sortBytes []byte

func (b sortBytes) Len() int {
	return len(b)
}
func (b sortBytes) Less(i, j int) bool {
	return b[i] < b[j]
}
func (b sortBytes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

type swapContext struct {
	size int
	goal [max_size]byte
}

type swapState struct {
	context *swapContext
	vector  [max_size]byte
	cost    float64
	op      int
}

func asSlice(s swapState) []byte {
	return s.vector[0:s.context.size]
}

func (s swapState) String() string {
	return fmt.Sprintf("%v, %d", asSlice(s), s.op)
}

func newSwapState(context *swapContext, vector [max_size]byte, cost float64, op int) swapState {
	return swapState{context, vector, cost, op}
}

func swapProblem(initialState []byte) swapState {
	if len(initialState) > max_size {
		panic(fmt.Sprintf("Maximum size of vector is %v, but found %v", max_size, len(initialState)))
	}
	var array [max_size]byte
	for i, v := range initialState {
		array[i] = v
	}
	sorted := array
	sort.Sort(sortBytes(sorted[0:len(initialState)]))
	context := &swapContext{len(initialState), sorted}
	return newSwapState(context, array, 0.0, -1)
}

// returns a copy of the given vector, where the element at index is swapped with its right neighbour
func swap(vector [max_size]byte, index int) [max_size]byte {
	vector[index], vector[index+1] = vector[index+1], vector[index]
	return vector
}

func (s swapState) Id() interface{} {
	return s.vector
}

func (s swapState) Expand() []solve.State {
	n := s.context.size - 1
	steps := make([]solve.State, n, n)
	for i := 0; i < n; i++ {
		steps[i] = newSwapState(s.context, swap(s.vector, i), s.cost+1.0, i)
	}
	return steps
}

func (s swapState) IsGoal() bool {
	return s.vector == s.context.goal
}

func (s swapState) Cost() float64 {
	return s.cost
}

func (s swapState) Heuristic() float64 {
	goal := s.context.goal
	n := s.context.size
	offset := 0
	for i := 0; i < n; i++ {
		value := s.vector[i]
		for d := 0; d < n; d++ {
			l, r := i-d, i+d
			if l >= 0 && goal[l] == value || r < n && goal[r] == value {
				offset += d
				break
			}
		}
	}
	return float64(offset / 2)
}

func printSolution(states []solve.State) {
	for _, state := range states {
		swapstate := state.(swapState)
		for i, e := range asSlice(swapstate) {
			if i > 0 {
				if i == swapstate.op+1 {
					fmt.Print("x")
				} else {
					fmt.Printf(" ")
				}
			}
			fmt.Print(e)
		}
		fmt.Println()
	}
}

func main() {
	state := swapProblem([]byte{7, 6, 5, 4, 3, 2, 1, 0})
	fmt.Printf("Sorting %v in minimal number of swaps of neighbouring elements\n", state)
	solution := solve.NewSolver(state).
		Algorithm(solve.IDAstar).
		Constraint(solve.CHEAPEST_PATH).
		Solve()

	fmt.Printf("visited: %d, expanded %d\n", solution.Visited, solution.Expanded)
	if len(solution.Solution) == 0 {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found in %v steps\n", len(solution.Solution))
		printSolution(solution.Solution)
	}
}
