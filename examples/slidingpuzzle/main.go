package main

import (
	"fmt"
	"strings"
	"math/rand"
	"github.com/bertbaron/solve"
	"math"
	"os"
	"log"
	"runtime/pprof"
)

const (
	N = 4
)

type direction byte

const (
	up direction = iota
	down direction = iota
	left direction = iota
	right direction = iota
)

func (d direction) String() string {
	switch d {
	case up: return "↑"
	case down: return "↓"
	case left: return "←"
	case right: return "→"
	}
	panic(fmt.Sprintf("Invalid direction: %d", d))
}

type puzzleContext struct {
	width    byte
	height   byte
	solution [N][N]byte
}

type puzzleState struct {
	context *puzzleContext
	board   [N][N]byte
	cost    int16
	x, y    byte
	dir     direction
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
	state.x, state.y = byte(width - 1), byte(height - 1)
	state.board[state.y][state.x] = 0
	state.context = &puzzleContext{byte(width), byte(height), state.board}
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
	for y := 0; y < int(p.context.height); y++ {
		values := make([]string, int(p.context.width))
		for x := 0; x < int(p.context.width); x++ {
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
	nw.board[p.y][p.x], nw.board[y][x] = nw.board[y][x], 0
	nw.x, nw.y, nw.dir = x, y, d
	return &nw
}

func shuffle(seed int64, p puzzleState, shuffles int) puzzleState {
	r := rand.New(rand.NewSource(seed))
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
	return float64(p.cost)
}

func (p puzzleState) Expand() []solve.State {
	children := make([]solve.State, 0)
	for d := 0; d < 4; d++ {
		if child := move(p, direction(d)); child != nil {
			child.cost += 1
			children = append(children, *child)
		}
	}
	return children
}

func (p puzzleState) IsGoal() bool {
	return p.board == p.context.solution
}

func manhattanDistance(w, h, x, y, value int) float64 {
	if value == 0 {
		return 0
	}
	xx, yy := value % w, value / w
	return math.Abs(float64(xx - x)) + math.Abs(float64(yy - y))
}

func (p puzzleState) Heuristic() float64 {
	md := 0.0
	w, h := int(p.context.width), int(p.context.height)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			md += (manhattanDistance(w, h, x, y, int(p.board[y][x])))
		}
	}
	return md
}

func (p puzzleState) Id() interface{} {
	return p.board
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	//seed := time.Now().UnixNano()
	var seed int64 = 3
	puzzle := shuffle(seed, initPuzzle(3, 3), 1000)
	fmt.Println("Solving the puzzle:")
	fmt.Print(puzzle.draw())
	fmt.Println()
	result := solve.NewSolver(puzzle).
		Algorithm(solve.IDAstar).
		Constraint(solve.NO_LOOP).
		Solve()
	n := len(result.Solution)
	if n == 0 {
		fmt.Println("No solution found")
	} else {
		moves := make([]string, n - 1)
		for i, state := range result.Solution[1:] {
			moves[i] = state.(puzzleState).dir.String()
		}
		fmt.Printf("Solution in %v steps: %s\n", result.Solution[n - 1].Cost(), strings.Join(moves, ", "))
		fmt.Printf("visited %d nodes\n", result.Visited)
	}
}