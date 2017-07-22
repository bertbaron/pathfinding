// sokoban solver, work in progress
package main

import (
	"fmt"
	"github.com/bertbaron/solve"
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

type context struct {
	world [][]byte
}

func (c context) Print() {
	for _, row := range c.world {
		for _, value := range row {
			fmt.Print(string(reverse[value]))
		}
		fmt.Println()
	}
}

type mainstate struct {
	boxes [][]byte
	x, y byte
	cost int
}

func (s mainstate) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s mainstate) Heuristic(ctx solve.Context) float64 {
	return 0
}

func (s mainstate) Expand(ctx solve.Context) []solve.State {
	return []solve.State{}
}

func (s mainstate) IsGoal(ctx solve.Context) bool {
	return false
}

func parse(level string) (context, mainstate) {
	var c context
	var s mainstate
	c.world = make([][]byte, 0)
	s.boxes = make([][]byte, 0)
	for y, row := range strings.Split(level, "\n") {
		worldrow := make([]byte, 0)
		boxrow := make([]byte, 0)
		for x, raw := range row {
			if value, ok := chars[raw]; ok {
				worldrow = append(worldrow, value)
				boxrow = append(boxrow, value|box)
				if value&player != 0 {
					s.x, s.y = byte(x), byte(y)
				}
			} else {
				panic(fmt.Sprintf("Invalid level format, character %v is not valid", value))
			}
		}
		if len(worldrow) > 0 {
			c.world = append(c.world, worldrow)
			s.boxes = append(s.boxes, boxrow)
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
	world.Print()
	result := solve.NewSolver(root).
		Context(world).
		Algorithm(solve.IDAstar).
		Solve()
	fmt.Printf("Result: %v\n ", result.Solution)
}
