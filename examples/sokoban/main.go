// sokoban solver, work in progress
package main

import (
	"fmt"
	"github.com/bertbaron/solve"
	"sort"
	"strings"
	"os"
	"log"
	"runtime/pprof"
	"time"
)

const (
	maxBoxes = 64
	unsolvable = 1000000
)

var simpleLevel = `
########
#     @#
#      #
# $ # .#
#  # # #
########`

var mediumLevel =`
   ####
####  ##
#      #
# $*.* #
#  *$.@##
## * .$ #
 ##.*.  #
  #$$ ###
  #   #
  #####`

var hardLevel = `
   ####
####  ##
#   $  #
#  *** #
#  . . ##
## * *  #
 ##***  #
  # $ ###
  # @ #
  #####`

const (
	floor  byte = 0
	wall   byte = 1
	box    byte = 2
	goal   byte = 4
	player byte = 8
)

var chars = map[rune]byte{
	' ': floor,
	'#': wall,
	'$': box,
	'.': goal,
	'@': player,
	'+': player | goal,
	'*': box | goal}

var reverse = map[byte]rune{
	floor:         ' ',
	wall:          '#',
	box:           '$',
	goal:          '·',
	player:        '@',
	player | goal: '+',
	box | goal:    '*'}

// -------- main problem. We only expose the states in which a block is pushed though to limit the search space
//          for the main search.
type sokoban struct {
	// the static world, without player and boxes
	world []byte
	// sorted list of goal positions
	goals  [maxBoxes]uint16
	boxcount int
	width  int
	height int
	// Shortes path from any position to nearest goal in a field without boxes
	heuristics []int
}

type mainstate struct {
	// sorted list of box positions
	boxes    [maxBoxes]uint16
	position int
	cost     int
	heuristic int
}

// returns the index of position in the sorted list of positions. Returns -1 if the position is not found
func binarySearch(positions []uint16, position int) int {
	idx := sort.Search(len(positions), func(i int) bool { return positions[i] >= uint16(position) })
	if idx < len(positions) && positions[idx] == uint16(position) {
		return idx
	}
	return -1
}

func valueOf(s sokoban, m *mainstate, position int) byte {
	boxidx := binarySearch(m.boxes[:s.boxcount], position)
	var additional byte = 0
	if m.position == position {
		additional |= player
	}
	if boxidx >= 0 {
		additional |= box
	}
	return s.world[position] | additional
}

func isEmpty(value byte) bool {
	return value&(wall|box) == 0
}

func isBox(value byte) bool {
	return value&box != 0
}

func isWall(value byte) bool {
	return value&wall != 0
}

func print(s sokoban, m *mainstate) {
	for position := range s.world {
		fmt.Print(string(reverse[valueOf(s, m, position)]))
		if position%s.width == s.width-1 {
			fmt.Println()
		}
	}
}

func (s *mainstate) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func boxHeuristic(world sokoban, box uint16) int {
	return world.heuristics[int(box)]
}

func totalHeuristic(world sokoban, s *mainstate) int {
	total := 0
	for _, box := range s.boxes[:world.boxcount] {
		total += boxHeuristic(world, box)
	}
	return total
}

func (s *mainstate) Heuristic(ctx solve.Context) float64 {
	return float64(s.heuristic)
}

var lastValue = -1 // TODO no global state!

func (s *mainstate) IsGoal(ctx solve.Context) bool {
	if s.cost + s.heuristic > lastValue {
		lastValue = s.cost + s.heuristic
		fmt.Printf("At depth %v (%v)\n", lastValue, s.heuristic)
	}
	for i, value := range ctx.Custom.(sokoban).goals {
		if s.boxes[i] != value {
			return false
		}
	}
	return true
}

func (s *mainstate) Expand(ctx solve.Context) []solve.State {
	world := ctx.Custom.(sokoban)
	targets := make([]int, 0)
	for _, box := range s.boxes[:world.boxcount] {
		left := isEmpty(valueOf(world, s, int(box)-1))
		right := isEmpty(valueOf(world, s, int(box)+1))
		up := isEmpty(valueOf(world, s, int(box)-world.width))
		down := isEmpty(valueOf(world, s, int(box)+world.width))
		if left && right {
			targets = append(targets, int(box)-1)
			targets = append(targets, int(box)+1)
		}
		if up && down {
			targets = append(targets, int(box)-world.width)
			targets = append(targets, int(box)+world.width)
		}
	}
	paths := getWalkMoves2(world, s, targets)
	children := make([]solve.State, 0, len(paths))
	for _, path := range paths {
		p := path.position
		for _, dir := range [...]int{-1, 1, -world.width, world.width} {
			if isBox(valueOf(world, s, p+dir)) && isEmpty(valueOf(world, s, p+2*dir)) {
				children = appendPushIfValid(children, world, s, p, dir, path.cost)
			}
		}
	}
	return children
}

func appendPushIfValid(children []solve.State, world sokoban, s *mainstate, position int, direction int, cost int) []solve.State {
	newposition := position + direction
	newbox := uint16(position + direction*2)
	newboxes := s.boxes
	idx := binarySearch(newboxes[:world.boxcount], newposition)
	newboxes[idx] = newbox

	n := world.boxcount
	// insertion sort to keep boxes sorted, only needed when moving up or down
	if direction < -1 {
		for idx > 0 && newboxes[idx-1] > newbox {
			newboxes[idx-1], newboxes[idx] = newboxes[idx], newboxes[idx-1]
			idx--
		}
	}
	if direction > 1 {
		for idx < n-1 && newboxes[idx+1] < newbox {
			newboxes[idx+1], newboxes[idx] = newboxes[idx], newboxes[idx+1]
			idx++
		}
	}
	newState := mainstate{newboxes, newposition, s.cost + cost + 1, 0}
	h := boxHeuristic(world, newbox)
	if h >= unsolvable || deadEnd(world, &newState, int(newbox)) {
		return children
	}
	newState.heuristic = s.heuristic - boxHeuristic(world, uint16(newposition)) + h
	return append(children, &newState)
}

// looks in a 3x3 pattern around the box position if this is a dead end
func deadEnd(world sokoban, s *mainstate, position int) bool {
	if world.world[position]&goal != 0 {
		return false // box is on a goal position
	}

	// corner walls
	lu := world.world[position-1-world.width]&wall != 0
	ru := world.world[position+1-world.width]&wall != 0
	ld := world.world[position-1+world.width]&wall != 0
	rd := world.world[position+1+world.width]&wall != 0

	// orthogonal walls or boxes
	uvalue := valueOf(world, s, position-world.width)
	dvalue := valueOf(world, s, position+world.width)
	lvalue := valueOf(world, s, position-1)
	rvalue := valueOf(world, s, position+1)

	// direction is blocked if it is a wall or a block that is sideways blocked by a wall
	u := isWall(uvalue) || (isBox(uvalue) && (lu || ru))
	d := isWall(dvalue) || (isBox(dvalue) && (ld || rd))
	l := isWall(lvalue) || (isBox(lvalue) && (lu || ld))
	r := isWall(rvalue) || (isBox(rvalue) && (ru || rd))

	return u && r || r && d || d && l || l && u
}

// -------------- Sub problem for moving the player to all positions in which a box can be moved -----------
/*
 This could be simply implemented with a breadth-first search using the library, but for this very simple
 problem which is executed on each expand in the main search the overhead of the library is significant.
 */
type walkstate struct {
	position int
	cost     int
}

func getWalkMoves2(wc sokoban, s *mainstate, targets []int) []walkstate {
	world := make([]byte, len(wc.world))
	copy(world, wc.world)
	for _, boxposition := range s.boxes {
		world[boxposition] = wall // mark boxes a wall
	}
	for _, target := range targets {
		world[target] = player // mark target positions as player
	}

	result := make([]walkstate, 0, len(targets))

	// breadth-first search
	queue := make([]walkstate, len(wc.world))
	queue[0] = walkstate{s.position, 0}
	world[s.position] |= wall
	start := 0
	end := 1
	for end > start {
		state := queue[start]
		start++
		value := world[state.position]

		if value & player != 0 {
			result = append(result, state)
		}
		for _, dir := range [...]int{-1,1,-wc.width, wc.width} {
			newPos := state.position + dir
			if world[newPos] & wall == 0 {
				world[newPos] |= wall // mark as enqueued by simply putting a wall there
				queue[end] = walkstate{newPos, state.cost+1}
				end++
			}
		}
	}
	return result
}


// ------------ Sub problem of moving a single box to its nearest target

type movestate struct {
	position int
	cost     int
}

func (s movestate) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s movestate) Heuristic(ctx solve.Context) float64 {
	return 0
}

func (s movestate) IsGoal(ctx solve.Context) bool {
	wc := ctx.Custom.(sokoban)
	return wc.world[s.position] & goal != 0
}

func (s movestate) Expand(ctx solve.Context) []solve.State {
	children := make([]solve.State, 0, 4)
	wc := ctx.Custom.(sokoban)

	p := s.position
	x, y := p % wc.width, p / wc.width
	if x==0 || x == wc.width-1 || y==0 || y==wc.height-1 {
		return children
	}

	left := isEmpty(wc.world[s.position-1])
	right := isEmpty(wc.world[s.position+1])
	up := isEmpty(wc.world[s.position-wc.width])
	down := isEmpty(wc.world[s.position+wc.width])
	if left && right {
		children = append(children, movestate{s.position-1, s.cost + 1})
		children = append(children, movestate{s.position+1, s.cost + 1})
	}
	if up && down {
		children = append(children, movestate{s.position-wc.width, s.cost + 1})
		children = append(children, movestate{s.position+wc.width, s.cost + 1})
	}
	return children
}

// TODO Efficiently share with walkstateMap
type movestateMap []float64

func (c movestateMap) Get(state solve.State) (float64, bool) {
	value := c[state.(movestate).position];
	return value, value >= 0
}

func (c movestateMap) Put(state solve.State, value float64) {
	c[state.(movestate).position] = value
}

func (c movestateMap) Clear() {
	for i := range c {
		c[i] = -1
	}
}

// Returns the cost of moving a box to its nearest goal
func costForMovingBlockToNearestGoal(wc sokoban, position int) int {
	//if wc.world[position] & goal != 0 {
	//	return 0 // small optimization
	//}
	rootstate := movestate{position, 0}
	msMap := make(movestateMap, len(wc.world))
	result := solve.NewSolver(rootstate).
		Context(wc).
		Constraint(solve.CheapestPathConstraint(msMap)).
		Algorithm(solve.BreadthFirst).
		Solve()
	if result.Solved() {
		return result.GoalState().(movestate).cost
	}
	return unsolvable
}



func parse(level string) (sokoban, *mainstate) {
	width := 0
	lines := strings.Split(level, "\n")
	height := len(lines)
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}
	var c sokoban
	var s mainstate
	c.width = width
	c.height = height

	c.world = make([]byte, width*height)
	boxcount := 0
	goalcount := 0
	for y, row := range lines {
		for x, raw := range row {
			position := y*width + x
			if value, ok := chars[raw]; ok {
				c.world[position] = value &^ player &^ box
				if value&player != 0 {
					s.position = position
				}
				if value&goal != 0 {
					c.goals[goalcount] = uint16(position)
					goalcount++
				}
				if value&box != 0 {
					s.boxes[boxcount] = uint16(position)
					boxcount++
				}
			} else {
				panic(fmt.Sprintf("Invalid level format, character %v is not valid", value))
			}
		}
	}
	c.boxcount = boxcount
	c.heuristics = make([]int, len(c.world))
	for i := range c.heuristics {
		c.heuristics[i] = costForMovingBlockToNearestGoal(c, i)
	}
	s.heuristic = totalHeuristic(c, &s)
	return c, &s
}

// For cheapest path constraint
type cpkey [maxBoxes+1]uint16
type cpMap map[cpkey]float64

func key(state solve.State) cpkey {
	var key cpkey
	s := state.(*mainstate)
	key[0] = uint16(s.position)
	copy(key[1:], s.boxes[:])
	return key
}

func (c cpMap) Get(state solve.State) (value float64, ok bool) {
	value, ok = c[key(state)]
	return
}

func (c cpMap) Put(state solve.State, value float64) {
	c[key(state)] = value
}

func (c *cpMap) Clear() {
	*c = make(cpMap)
}

func cheapestPathConstraint() solve.Constraint {
	var m cpMap
	return solve.CheapestPathConstraint(&m)
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	world, root := parse(mediumLevel)
	print(world, root)
	start := time.Now()
	result := solve.NewSolver(root).
		Context(world).
		Algorithm(solve.Astar).
		Constraint(cheapestPathConstraint()).
		//Limit(44).
		Solve()
	fmt.Printf("Time: %.1f seconds\n", time.Since(start).Seconds())
	if result.Solved() {
		fmt.Printf("Result:\n ")
		//for _, state := range result.Solution {
		//	print(world, state.(*mainstate))
		//}
		fmt.Printf("Solved in %d moves\n", int(result.GoalState().(*mainstate).cost))
	}
	fmt.Printf("visited %v and expanded %v main nodes\n", result.Visited, result.Expanded)

	f, err = os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
	return
}
