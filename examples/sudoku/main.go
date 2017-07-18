// Very simple sudoku solver to show the use of depth-first
// to solve a problem with plain simple backtracking
package main

import (
	"fmt"
	"github.com/bertbaron/solve"
)

type sudoku struct {
	values   [9][9]int
	position int
}

func newSudoku(values [9][9]int) sudoku {
	return sudoku{values, 0}
}

func (s sudoku) Cost(solve.Context) float64 {
	return 0
}

func (s sudoku) Heuristic(solve.Context) float64 {
	return 0
}

func (s sudoku) IsGoal(solve.Context) bool {
	return s.position == 9*9
}

func (s sudoku) Expand(solve.Context) []solve.State {
	var children []solve.State
	for v := 1; v <= 9; v++ {
		if child := s.withValue(v); child != nil {
			children = append(children, *child)
		}
	}
	return children
}

// Returns a new sudoku with the value set at the current position,
// or nil if that would be invalid
func (s sudoku) withValue(value int) *sudoku {
	row := s.position / 9
	col := s.position % 9
	copy := s
	copy.position++
	current := s.values[row][col]
	if current == value {
		return &copy
	}
	if current != 0 {
		return nil
	}
	for i := 0; i < 9; i++ {
		if s.values[row][i] == value || s.values[i][col] == value || s.values[row/3*3+i/3][col/3*3+i%3] == value {
			return nil // row, column or block conflict
		}
	}
	copy.values[row][col] = value
	return &copy
}

func (s sudoku) Print() {
	for _, row := range s.values {
		for _, value := range row {
			if value == 0 {
				fmt.Print("  ")
			} else {
				fmt.Printf(" %d", value)
			}
		}
		fmt.Println()
	}
}

func main() {
	s := newSudoku([9][9]int{
		{0, 0, 6, 9, 0, 5, 0, 0, 2},
		{0, 4, 0, 0, 0, 0, 8, 0, 0},
		{0, 0, 0, 0, 1, 0, 0, 4, 5},
		{0, 8, 0, 6, 0, 4, 7, 0, 0},
		{0, 0, 0, 0, 2, 0, 0, 0, 9},
		{7, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 2, 0, 0},
		{0, 0, 1, 0, 0, 7, 0, 6, 0},
		{0, 3, 0, 4, 6, 0, 0, 0, 1}})
	fmt.Println("Solving:")
	s.Print()
	solver := solve.NewSolver(s).
		Algorithm(solve.DepthFirst)
	result := solver.Solve()
	if !result.Solved() {
		fmt.Println("No solution found")
	} else {
		fmt.Println("Solution:")
		result.GoalState().(sudoku).Print()

		if solver.Solve().Solved() {
			fmt.Println("There is more than 1 solution for this sudoku")
		}
	}
}
