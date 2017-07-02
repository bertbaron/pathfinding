package main

import (
	"fmt"
	"strings"
	"math/rand"
	"time"
	"github.com/bertbaron/solve"
)

const (
	N = 8
)

type direction int

const (
	up direction = iota
	down direction = iota
	left direction = iota
	right direction = iota
)

func (d direction) String() string {
	switch d {
	case up: return "UP"
	case down: return "DOWN"
	case left: return "LEFT"
	case right: return "RIGHT"
	}
	panic(fmt.Sprintf("Invalid direction: %d", d))
}

type puzzleContext struct {
	width    int
	height   int
	solution [N][N]byte
}

type puzzleState struct {
	context *puzzleContext
	board   [N][N]byte
	x, y    int
	cost    float64
}

func initPuzzle(width, height int) puzzleState {
	var state puzzleState
	var value byte
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value++
			state.board[y][x] = value
		}
	}
	state.x = width - 1
	state.y = height - 1
	state.board[state.y][state.x] = 0
	state.context = &puzzleContext{width, height, state.board}
	return state
}

func byte2string(b byte) string {
	if b == 0 {
		return "  "
	}
	return fmt.Sprintf("%2d", b)
}

func (p puzzleState) draw() string {
	s := ""
	for y := 0; y < p.context.height; y++ {
		values := make([]string, p.context.width)
		for x := 0; x < p.context.height; x++ {
			values[x] = byte2string(p.board[y][x])
		}
		s += strings.Join(values, " ") + "\n"
	}
	return s
}

func move(p puzzleState, d direction) *puzzleState {
	x, y := p.x, p.y
	switch d {
	case up: y--
	case down: y++
	case left: x--
	case right: x++
	}
	if x < 0 || y < 0 || x >= p.context.width || y >= p.context.height {
		return nil
	}
	nw := p
	nw.board[p.y][p.x] = nw.board[y][x]
	nw.board[y][x] = 0
	nw.x = x
	nw.y = y
	return &nw
}

func shuffle(p puzzleState, shuffles int) puzzleState {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < shuffles; i++ {
		dir := direction(r.Intn(4))
		if nw := move(p, dir); nw != nil {
			p = *nw
		}

	}
	return p
}

/*
 Implementation of solve.State
 */

func (p puzzleState) Cost() float64 {
	return p.cost
}

func (p puzzleState) Expand() []solve.State {
	children := make([]solve.State, 0)
	for d := 0; d < 4; d++ {
		if child := move(p, direction(d)); child != nil {
			child.cost = p.cost + 1
			children = append(children, *child)
		}
	}
	return children
}

func (p puzzleState) IsGoal() bool {
	return p.board == p.context.solution
}

func (p puzzleState) Heuristic() float64 {
	return 0
}

func (p puzzleState) Id() interface{} {
	return p.board
}

func main() {
	puzzle := shuffle(initPuzzle(3, 3), 1000)
	fmt.Println("Solving the puzzle:")
	fmt.Print(puzzle.draw())
	fmt.Println()
	result := solve.NewSolver(puzzle).
		Algorithm(solve.IDAstar).
		Constraint(solve.CHEAPEST_PATH).
		Solve()
	fmt.Printf("Result: %v", result)
}