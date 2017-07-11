package solve_test

import (
	"fmt"
	"github.com/bertbaron/solve"
)

var puzzle solve.State

func ExampleSolver() {
	result := solve.NewSolver(puzzle).
		Algorithm(solve.IDAstar).
		Constraint(solve.NoLoopConstraint(10, sameState)).
		Solve()
	fmt.Println(result.Solution)
}
