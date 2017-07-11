package main

import (
	"fmt"
	"github.com/bertbaron/solve"
	"math/rand"
	"strings"
	"time"
)

const (
	height = 4
	width = 4
)

type direction int8

const (
	left direction = iota
	up direction = iota
	down direction = iota
	right direction = iota
)

func (d direction) String() string {
	switch d {
	case left:
		return "←"
	case up:
		return "↑"
	case right:
		return "→"
	case down:
		return "↓"
	}
	panic(fmt.Sprintf("Invalid direction: %d", d))
}

type subState struct {
	board [height][width]byte
	cost  int16
}

func (p subState) Cost(context *interface{}) float64 {
	return float64(p.cost)
}

func mkchild(children []solve.State, p subState, x, y, nx, ny int) []solve.State {
	if p.board[ny][nx] != 0 {
		return children
	}
	child := p
	child.board[ny][nx], child.board[y][x] = child.board[y][x], 0
	child.cost++
	return append(children, child)
}

func (p subState) Expand(context *interface{}) []solve.State {
	children := make([]solve.State, 0)
	for y, row := range p.board {
		for x, value := range row {
			if value != 0 {
				if y > 0 {
					children = mkchild(children, p, x, y, x, y - 1)
				}
				if y < width - 1 {
					children = mkchild(children, p, x, y, x, y + 1)
				}
				if x > 0 {
					children = mkchild(children, p, x, y, x - 1, y)
				}
				if x < height - 1 {
					children = mkchild(children, p, x, y, x + 1, y)
				}
			}
		}
	}
	return children
}

func (p subState) IsGoal(context *interface{}) bool {
	return isGoal(p.board)
}

func abs(value int) int {
	if (value < 0) {
		return -value
	}
	return value
}

func (p subState) Heuristic(context *interface{}) float64 {
	return float64(manhattanWithConflicts(p.board))
}

func (p subState) Id() interface{} {
	return p.board
}

func toSubstate(puzzle puzzleState, pattern [width * height]int) subState {
	var substate subState
	substate.board = puzzle.board
	for y, row := range substate.board {
		for x, value := range row {
			if pattern[value] == 1 {
				substate.board[y][x] = 0
			}
		}
	}
	return substate
}

var db = make(map[[height][width]byte]int)

func subHeuristic(puzzle puzzleState, pattern [width * height]int) int {
	substate := toSubstate(puzzle, pattern)
	if cached, ok := db[substate.board]; ok {
		//fmt.Print(".")
		return cached
	}
	//fmt.Printf("Substate: %v\n", substate)
	result := solve.NewSolver(substate).
	//Algorithm(solve.IDAstar).
		Algorithm(solve.Astar).
		Constraint(solve.CHEAPEST_PATH).
		Solve()
	n := len(result.Solution)
	if n == 0 {
		panic("Geen oplossing gevonden voor subprobleem!")
	}
	h := int(result.Solution[n - 1].Cost(nil))
	//fmt.Printf("Heuristic for %v: %v\n", substate, h)
	for _, state := range result.Solution {
		db[state.(subState).board] = h
	}
	//fmt.Print("x")
	return h
}

func subproblemHeuristic(puzzle puzzleState) int {
	h := 0
	h += subHeuristic(puzzle, [width * height]int{
		1, 1, 1, 1,
		1, 0, 0, 0,
		1, 0, 0, 0,
		1, 0, 0, 0})
	h += subHeuristic(puzzle, [width * height]int{
		0, 0, 0, 0,
		0, 1, 1, 1,
		0, 1, 1, 1,
		0, 1, 1, 1})
	return h
}

type puzzleState struct {
	board [height][width]byte
	cost  int16
	x, y  byte
	dir   direction
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
	return state
}

func fromBoard(board [][]int) puzzleState {
	var state puzzleState
	for y, row := range board {
		for x, value := range row {
			state.board[y][x] = byte(value)
			if value == 0 {
				state.x = byte(x)
				state.y = byte(y)
			}
		}
	}
	width := len(board[0])
	height := len(board)
	initPuzzle(width, height) // initializes the context as side-effect
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
	for y := 0; y < height; y++ {
		values := make([]string, width)
		for x := 0; x < width; x++ {
			values[x] = byte2string(p.board[y][x])
		}
		s += strings.Join(values, " ") + "\n"
	}
	return s
}

func move(p puzzleState, d direction) *puzzleState {
	x, y := p.x, p.y
	switch d {
	case up:
		y--
	case down:
		y++
	case left:
		x--
	case right:
		x++
	}
	if x < 0 || y < 0 || x >= width || y >= height {
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

func (p puzzleState) Cost(context *interface{}) float64 {
	return float64(p.cost)
}

func (p puzzleState) Expand(context *interface{}) []solve.State {
	children := make([]solve.State, 0)
	for d := 0; d < 4; d++ {
		if int(p.dir) != 3 - d {
			if child := move(p, direction(d)); child != nil {
				child.cost += 1
				children = append(children, *child)
			}
		}
	}
	return children
}

func isGoal(board [height][width]byte) bool {
	for y, row := range board {
		for x, value := range row {
			if value != 0 {
				if value != byte(y * width + x + 1) {
					return false
				}
			}
		}
	}
	return true
}

func (p puzzleState) IsGoal(context *interface{}) bool {
	return isGoal(p.board)
}

func manhattanWithConflicts(board [height][width]byte) int {
	heuristic := 0

	// manhattan distance + horizontal and vertical conflicts in single pass
	var maxvert [width]int
	for y, row := range board {
		maxhor := 0
		for x, value := range row {
			v := int(value)
			if v != 0 {
				xx, yy := (v - 1) % width, (v - 1) / width
				heuristic += abs(xx - x) + abs(yy - y)
				if yy == y {
					if (v > maxhor) {
						maxhor = v
					} else {
						heuristic += 2
					}
				}
				if xx == x {
					if (v > maxvert[x]) {
						maxvert[x] = v
					} else {
						heuristic += 2
					}
				}
			}
		}
	}
	return heuristic
}

func (p puzzleState) Heuristic(context *interface{}) float64 {
	if p.cost < 40 {
		return float64(subproblemHeuristic(p))
	}
	return float64(manhattanWithConflicts(p.board))
}

func (p puzzleState) Id() interface{} {
	return p.board
}

func generateAndSolve(seed int64) solve.Result {
	//puzzle := shuffle(seed, initPuzzle(4, 4), 10000)
	//fmt.Printf("Solving the puzzle generated with seed %v\n", seed)
	puzzle := fromBoard([][]int{{15, 14, 8, 12}, {10, 11, 9, 13}, {2, 6, 5, 1}, {3, 7, 4, 0}}) // 80 moves
	fmt.Print(puzzle.draw())
	fmt.Println()
	fmt.Printf("Heuristic: %v\n", manhattanWithConflicts(puzzle.board))
	fmt.Printf("Subproblem heuristic: %v\n", subproblemHeuristic(puzzle))
	fmt.Println()
	start := time.Now()
	result := solve.NewSolver(puzzle).
		Algorithm(solve.IDAstar).
	Limit(1).
		Solve()
	fmt.Printf("Time: %.2f\n", time.Since(start).Seconds())
	return result
	//n := len(result.Solution)
	//if n == 0 {
	//	fmt.Println("No solution found")
	//} else {
	//	moves := make([]string, n - 1)
	//	for i, state := range result.Solution[1:] {
	//		moves[i] = state.(puzzleState).dir.String()
	//	}
	//	fmt.Printf("Solution in %v steps: %s\n", result.Solution[n - 1].Cost(), strings.Join(moves, ", "))
	//	fmt.Printf("visited %d, expanded %d\n", result.Visited, result.Expanded)
	//}
}

func main() {
	//worstSeed := -1
	//worstVisited := 0
	//for seed :=0; seed <100; seed++ {
	//	result := generateAndSolve(int64(seed))
	//	if result.Visited > worstVisited {
	//		worstSeed = seed
	//		worstVisited = result.Visited
	//	}
	//	fmt.Printf("Worst: %v (%v)\n", worstSeed, worstVisited)
	//}

	//f, err := os.Create("cpu.prof")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	//
	result := generateAndSolve(8)
	n := len(result.Solution)
	if n == 0 {
		fmt.Println("No solution found")
	} else {
		moves := make([]string, n - 1)
		for i, state := range result.Solution[1:] {
			moves[i] = state.(puzzleState).dir.String()
		}
		fmt.Printf("Solution in %v steps: %s\n", len(result.Solution) - 1, strings.Join(moves, ", "))
		fmt.Printf("visited %d, expanded %d\n", result.Visited, result.Expanded)
	}
}
