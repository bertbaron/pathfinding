package main

import (
	"fmt"
	"github.com/bertbaron/solve"
	"sort"
	"time"
	"os"
	"log"
	"runtime/pprof"
)

const maxSize = 16

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
	goal [maxSize]byte
}

type swapState struct {
	vector  [maxSize]byte
	cost    float64
	op      int
}

func (s swapState) String() string {
	return fmt.Sprintf("%v, %d", s.vector, s.op)
}

func newSwapState(vector [maxSize]byte, cost float64, op int) swapState {
	return swapState{vector, cost, op}
}

func swapProblem(initialState []byte) (swapContext, swapState) {
	if len(initialState) > maxSize {
		panic(fmt.Sprintf("Maximum size of vector is %v, but found %v", maxSize, len(initialState)))
	}
	var array [maxSize]byte
	for i, v := range initialState {
		array[i] = v
	}
	sorted := array
	sort.Sort(sortBytes(sorted[0:len(initialState)]))
	context := swapContext{len(initialState), sorted}
	return context, newSwapState(array, 0.0, -1)
}

// returns a copy of the given vector, where the element at index is swapped with its right neighbour
func swap(vector [maxSize]byte, index int) [maxSize]byte {
	vector[index], vector[index+1] = vector[index+1], vector[index]
	return vector
}

func context(ctx solve.Context) swapContext {
	return (ctx.Custom).(swapContext)
}

func (s swapState) Id() interface{} {
	return s.vector
}

func (s swapState) Expand(ctx solve.Context) []solve.State {
	n := context(ctx).size - 1
	steps := make([]solve.State, n, n)
	for i := 0; i < n; i++ {
		steps[i] = newSwapState(swap(s.vector, i), s.cost+1.0, i)
	}
	return steps
}

func (s swapState) IsGoal(ctx solve.Context) bool {
	return s.vector == context(ctx).goal
}

func (s swapState) Cost(ctx solve.Context) float64 {
	return s.cost
}

func (s swapState) Heuristic(ctx solve.Context) float64 {
	return 0
	goal := context(ctx).goal
	n := context(ctx).size
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

func printSolution(context swapContext, states []solve.State) {
	for _, state := range states {
		swapstate := state.(swapState)
		for i :=0; i<context.size; i++ {
			e := state.(swapState).vector
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

type cpMap map[[maxSize]byte]solve.ConstraintNode

func (c cpMap) Get(state solve.State) (solve.ConstraintNode, bool) {
	value, ok := c[state.(swapState).vector]
	return value, ok
}

func (c cpMap) Put(state solve.State, value solve.ConstraintNode) {
	c[state.(swapState).vector] = value
}
func (c *cpMap) Clear() {
	*c = make(cpMap)
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	context, state := swapProblem([]byte{7, 6, 5, 4, 3, 2, 1, 0})
	fmt.Printf("Sorting %v in minimal number of swaps of neighbouring elements\n", state)
	constraintMap := make(cpMap)
	start := time.Now()
	solution := solve.NewSolver(state).
		Context(context).
		Algorithm(solve.IDAstar).
		//Constraint(solve.CheapestPathConstraint()).
		Constraint(solve.CheapestPathConstraint2(&constraintMap)).
		Solve()

	fmt.Printf("visited: %d, expanded %d, time %0.2fs\n", solution.Visited, solution.Expanded, time.Since(start).Seconds())
	if len(solution.Solution) == 0 {
		fmt.Printf("No solution found\n")
	} else {
		fmt.Printf("Solution found in %v steps\n", len(solution.Solution))
		//printSolution(solution.Solution)
	}
}
