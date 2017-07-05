package solve_test

import (
	"github.com/bertbaron/solve"
	"fmt"
)

var puzzle solve.State

func ExampleSolver() {
	result := solve.NewSolver(puzzle).
		Algorithm(solve.IDAstar).
		Constraint(solve.NO_LOOP).
		Solve()
	fmt.Println(result.Solution)
}
