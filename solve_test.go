package solve

import (
	"unicode"
	"testing"
	"math"
	"fmt"
)

/*
 Problem for testing modelled as a graph. The root node is always "a", and any node that starts with an uppercase
 character is a goal.
 */

type edge struct {
	target string
	cost   float64
}
type graph map[string][]edge
type state struct {
	graph graph
	node  string
	cost  float64
}

func create(graph graph) state {
	return state{graph, "a", 0.0}
}
func expand(s state, edge edge) state {
	return state{s.graph, edge.target, s.cost + edge.cost}
}

func (s state) Cost() float64 {
	return s.cost
}

func (s state) IsGoal() bool {
	return unicode.IsUpper([]rune(s.node)[0])
}

func (s state) Expand() []State {
	children := make([]State, 0)
	if edges, ok := s.graph[s.node]; ok {
		for _, edge := range edges {
			children = append(children, expand(s, edge))
		}
	}
	return children
}

func (s state) Heuristic() float64 {
	return 0
}

func (s state) Id() interface{} {
	return s.node
}

func testSolve(t *testing.T, graph graph, algorithm Algorithm, constraint Constraint, limit float64, solution string, costs float64) {
	result := NewSolver(create(graph)).
		Algorithm(algorithm).
		Constraint(constraint).
		Limit(limit).
		Solve()

	name := fmt.Sprintf("(%v,%v)", algorithm, constraint)
	if len(result.Solution) == 0 {
		t.Errorf("%v - Expected %v, but no solution found", name, solution)
		return
	}
	actual := result.Solution[len(result.Solution) - 1]
	state := actual.(state)
	if state.node != solution {
		t.Errorf("%v - Expected %v, but found %v", name, solution, state.node)
		return
	}
	if state.cost != costs {
		t.Errorf("%v - Expected cost %v, but found %v", name, costs, state.cost)
		return
	}
}

// Solves the problem with all algorithms and constraints that should return in the optimal solution
func testSolveAllAlgorithms(t *testing.T, graph graph, includeBF bool, solution string, costs float64) {
	testSolve(t, graph, Astar, NONE, math.MaxFloat64, solution, costs)
	testSolve(t, graph, Astar, NO_RETURN, math.MaxFloat64, solution, costs)
	testSolve(t, graph, Astar, NO_LOOP, math.MaxFloat64, solution, costs)
	testSolve(t, graph, Astar, CHEAPEST_PATH, math.MaxFloat64, solution, costs)

	testSolve(t, graph, IDAstar, NONE, math.MaxFloat64, solution, costs)
	testSolve(t, graph, IDAstar, NO_RETURN, math.MaxFloat64, solution, costs)
	testSolve(t, graph, IDAstar, NO_LOOP, math.MaxFloat64, solution, costs)
	testSolve(t, graph, IDAstar, CHEAPEST_PATH, math.MaxFloat64, solution, costs)

	testSolve(t, graph, DepthFirst, NONE, costs, solution, costs)
	testSolve(t, graph, DepthFirst, NO_RETURN, costs, solution, costs)
	testSolve(t, graph, DepthFirst, NO_LOOP, costs, solution, costs)
	testSolve(t, graph, DepthFirst, CHEAPEST_PATH, costs, solution, costs)

	// BF is only optimal if the length of costs corresonds with the length of the path
	if includeBF {
		testSolve(t, graph, BreadthFirst, NONE, math.MaxFloat64, solution, costs)
		testSolve(t, graph, BreadthFirst, NO_RETURN, math.MaxFloat64, solution, costs)
		testSolve(t, graph, BreadthFirst, NO_LOOP, math.MaxFloat64, solution, costs)
		testSolve(t, graph, BreadthFirst, CHEAPEST_PATH, math.MaxFloat64, solution, costs)
	}
}

func TestSimpleProblem(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"b", 1}, {"c", 1}}
	g["b"] = []edge{{"D", 1}, {"c", 1}}
	testSolveAllAlgorithms(t, g, true, "D", 2)
}

func TestOptimalEvenIfPathLooksBad(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"b", 1}, {"c", 8}, {"d", 10}}
	g["b"] = []edge{{"bb", 1}}
	g["c"] = []edge{{"cc", 8}}
	g["d"] = []edge{{"dd", 10}}
	g["bb"] = []edge{{"B", 200}}
	g["cc"] = []edge{{"C", 100}}
	g["dd"] = []edge{{"D", 1}}
	testSolveAllAlgorithms(t, g, false, "D", 21)
}

func TestIDAStarWithInfinitContour(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"b", math.Inf(1)}}

	result := NewSolver(create(g)).
		Algorithm(IDAstar).
		Solve()
	if len(result.Solution) != 0 {
		t.Error("Expected no solution, but found one")
	}
}