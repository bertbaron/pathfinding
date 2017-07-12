package main

import (
	"github.com/bertbaron/solve"
	"fmt"
)

type sudoku struct {
	values   [9][9]int
	position int
}

func (s sudoku) String() string {
	return fmt.Sprintf("%v (%v)", s.values, s.position)
}

func (s sudoku) Cost(solve.Context) float64 {
	return float64(s.position)
}

func (s sudoku) Heuristic(solve.Context) float64 {
	return 0
}

var max int
func (s sudoku) IsGoal(solve.Context) bool {
	if s.position > max {
		max = s.position
		fmt.Println(s.values)
	}
	return s.position == 9 * 9
}

func (s sudoku) Expand(solve.Context) []solve.State {
	var children []solve.State
	for v := 1; v < 10; v++ {
		if child := s.withValue(v); child != nil {
			children = append(children, *child)
		}

	}
	return children
}

func (s sudoku) withValue(value int) *sudoku {
	row := s.position / 9
	col := s.position % 9
	current := s.values[row][col]
	if current == value {
		copy := s
		copy.position++
		return &copy
	}
	if current != 0 {
		return nil
	}
	blockx := col / 3 * 3
	blocky := row / 3 * 3
	for i:=0; i<9; i++ {
		bx := blockx + i % 3
		by := blocky + i / 3
		if s.values[row][i] == value || s.values[i][col] == value || s.values[by][bx] == value {
			return nil
		}
	}
	copy := s
	copy.values[row][col] = value
	copy.position++
	return &copy
}

func main() {
	var s sudoku
	result := solve.NewSolver(s).
		Algorithm(solve.DepthFirst).
		Solve()
	fmt.Println("Visited:", result.Visited)
	n := len(result.Solution)
	if n == 0 {
		fmt.Println("No solution found")
	} else {
		fmt.Printf("Solution: %v \n", result.Solution[n-1].(sudoku).values)
	}
}