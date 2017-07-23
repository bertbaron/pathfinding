// sokoban solver, work in progress
package main

import (
	"fmt"
	"github.com/bertbaron/solve"
	"sort"
	"strings"
)

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
	goal:          '.',
	player:        '@',
	player | goal: '+',
	box | goal:    '*'}

// -------- main problem. We only expose the states in which a block is pushed though to limit the search space
//          for the main search.
type sokoban struct {
	// the static world, without player and boxes
	world []byte
	// sorted list of goal positions
	goals  []uint16
	width  int
	height int
}

type mainstate struct {
	// sorted list of box positions
	boxes    []uint16
	position int
	cost     int
}

func valueOf(s *sokoban, m *mainstate, position int) byte {
	boxidx := sort.Search(len(m.boxes), func(i int) bool { return m.boxes[i] >= uint16(position) })
	var additional byte = 0
	if m.position == position {
		additional |= player
	}
	if boxidx < len(m.boxes) && m.boxes[boxidx] == uint16(position) {
		additional |= box
	}
	return s.world[position] | additional
}

func print(s sokoban, m mainstate) {
	for position := range s.world {
		fmt.Print(string(reverse[valueOf(&s, &m, position)]))
		if position%s.width == s.width-1 {
			fmt.Println()
		}
	}
}

func (s mainstate) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s mainstate) Heuristic(ctx solve.Context) float64 {
	return 0
}

func (s mainstate) IsGoal(ctx solve.Context) bool {
	for i, value := range ctx.Custom.(sokoban).goals {
		if s.boxes[i] != value {
			return false
		}
	}
	return true
}

func (s mainstate) Expand(ctx solve.Context) []solve.State {
	var children []solve.State
	return children
}

// -------------- Sub problem for moving the player to all positions in which a box can be moved -----------

type walkcontext struct {
	// the static world, without player but with boxes because we don't move them here
	world []byte
	goalpositions []int
	width int
}

type walkstate struct {
	position int
	cost int
}

func (s walkstate) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s walkstate) Heuristic(ctx solve.Context) float64 {
	return 0
}

func (s walkstate) IsGoal(ctx solve.Context) bool {
	wc := ctx.Custom.(walkcontext)
	for _, goal := range wc.goalpositions {
		if s.position == goal {
			return true
		}
	}
	return false
}

func (s walkstate) Expand(ctx solve.Context) []solve.State {
	var children []solve.State
	wc := ctx.Custom.(walkcontext)
	children = s.addIfValid(children, s.position-1, wc)
	children = s.addIfValid(children, s.position+1, wc)
	children = s.addIfValid(children, s.position-wc.width, wc)
	children = s.addIfValid(children, s.position+wc.width, wc)
	return children
}

func (s walkstate) addIfValid(children []solve.State, newPosition int, wc walkcontext) []solve.State {
	if wc.world[newPosition] & (wall | box) == 0 {
		return append(children, walkstate{newPosition, s.cost + 1})
	}
	return children
}

func parse(level string) (sokoban, mainstate) {
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
	c.goals = make([]uint16, 0)
	s.boxes = make([]uint16, 0)
	for y, row := range lines {
		for x, raw := range row {
			position := y*width + x
			if value, ok := chars[raw]; ok {
				c.world[position] = value &^ player &^ box
				if value&player != 0 {
					s.position = position
				}
				if value&goal != 0 {
					c.goals = append(c.goals, uint16(position))
				}
				if value&box != 0 {
					s.boxes = append(s.boxes, uint16(position))
				}
			} else {
				panic(fmt.Sprintf("Invalid level format, character %v is not valid", value))
			}
		}
	}
	return c, s
}

var level = `
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

func main() {
	world, root := parse(level)
	print(world, root)
	result := solve.NewSolver(root).
		Context(world).
		Algorithm(solve.IDAstar).
		Solve()
	fmt.Printf("Result: %v\n ", result.Solution)
}
