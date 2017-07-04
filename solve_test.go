package solve

import (
	"unicode"
	"testing"
	"math"
	"fmt"
	"math/rand"
)

/*
 Problem for testing modelled as a graph. The root node is always "a" or "A", and any node that starts with an uppercase
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
func (s state) String() string {
	return fmt.Sprintf("%v", s.node)
}

func create(graph graph) state {
	var root = "a"
	if _, ok := graph[root]; !ok {
		root = "A"
	}
	return state{graph, root, 0.0}
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

func TestIDAStarWithMaxFlaotContour(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"b", math.MaxFloat64}}

	result := NewSolver(create(g)).
		Algorithm(IDAstar).
		Solve()
	if len(result.Solution) != 0 {
		t.Error("Expected no solution, but found one")
	}
}

func TestWithSingleStateResult(t *testing.T) {
	g := make(graph)
	g["A"] = []edge{}
	result := NewSolver(create(g)).
		Algorithm(IDAstar).
		Solve()
	if (len(result.Solution) != 1) {
		t.Errorf("Expected solution in one step, but found %v", len(result.Solution))
	}
}

func testStatistics(t *testing.T, g graph, algorithm Algorithm, constraint Constraint, expExpanded, expVisited int) {
	name := fmt.Sprintf("(%v,%v)", algorithm, constraint)
	result := NewSolver(create(g)).
		Algorithm(algorithm).
		Constraint(constraint).
		Solve()
	if result.Visited != expVisited {
		t.Errorf("%v - Expected %v nodes visited, but was %v", name, expVisited, result.Visited)
	}
	if result.Expanded != expExpanded {
		t.Errorf("%v - Expected %v nodes expanded, but was %v", name, expExpanded, result.Expanded)
	}
}
/*
(deftest test-with-no-loop-constraint
  (let [graph {:a  [[:a 1] [:b 1]]
               :b  [[:c 1] [:d 2]]
               :c  [[:a 1] [:d 1]]
               :d  [[:E 1]]}]
    (test-statistics graph :A* {:expanded 22 :visited 12})
    (test-statistics graph :A* {:expanded  8 :visited  6} :constraint (no-return-constraint))
    (test-statistics graph :A* {:expanded  6 :visited  5} :constraint (no-loop-constraint))))
 */
func TestStatisticsWithDifferentConstraints(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"a", 1}, {"b", 1}}
	g["b"] = []edge{{"c", 1}, {"d", 2}}
	g["c"] = []edge{{"a", 1}, {"d", 1}}
	g["d"] = []edge{{"E", 1}}
	testStatistics(t, g, Astar, NONE, 27, 16) // currently the test depends on the queue implementation...
	testStatistics(t, g, Astar, NO_RETURN, 8, 7)
	testStatistics(t, g, Astar, NO_LOOP, 6, 6)
	testStatistics(t, g, Astar, CHEAPEST_PATH, 4, 5)
}

type dummyState struct {
	State
	name string
}

func (s dummyState) Id() interface{} {
	return s.name
}

func dummyNode(parent *node, name string, costs float64) *node {
	return &node{parent, dummyState{nil, name}, costs}
}

func TestNoLoopConstraint(t *testing.T) {
	assert := func (name string, value, expected interface {}) {
		if value != expected {
			t.Errorf("%v - Expected %v, but was %v", name, expected, value)
		}
	}

	c := noLoopConstraint(2)
	a1 := dummyNode(nil, "a", 1)
	assert("a1", c.onExpand(a1), false)
	a2 := dummyNode(a1, "a", 1)
	assert("same parent", c.onExpand(a2), true)


	b1 := dummyNode(a1, "b", 1)
	assert("b1", c.onExpand(b1), false)

	// a - b - a
	a3 := dummyNode(b1, "a", 1)
	assert("same grandparent", c.onExpand(a3), true)

	c1 := dummyNode(b1, "c", 1)
	assert("c1", c.onExpand(c1), false)

	// a - b - c - a
	a4 := dummyNode(c1, "a", 1)
	assert("same grandgrandparent", c.onExpand(a4), false)
}

func TestRingbuffer(t *testing.T) {
	mknode := func(i int) *node {
		return &node{nil, nil, float64(i)}
	}
	b := breadthFirst()
	lastTaken := -1
	for i:=0; i<1000; i++ {
		b.Add(mknode(i))
		if i % 3 == 0 {
			taken := b.Take()
			if taken == nil {
				t.Errorf("Expected node %v at head of the buffer, but the buffer was empty", lastTaken + 1)
				return
			}
			if (int(taken.value) != lastTaken + 1) {
				t.Errorf("Expected element %v from the buffer, but was %v", lastTaken + 1, taken.value)
				return
			}
			lastTaken = int(taken.value)
		}
	}
}

func BenchmarkBreadthFirstStrategy(b *testing.B) {
	// for breadthfirst we can reuse the node, reducing overhead
	node := &node{nil, nil, 0}
	for n := 0; n < b.N; n++ {
		b := breadthFirst()
		for i:=0; i<3000000; i++ {
			b.Add(node)
			if i % 3 == 0 {
				b.Take()
			}
		}
	}
}

func BenchmarkAStarStrategy(b *testing.B) {
	// for Astar we can not reuse the node, so this test involves more overhead
	mknode := func(value float64) *node {
		return &node{nil, nil, value}
	}

	r := rand.New(rand.NewSource(123))
	for n := 0; n < b.N; n++ {
		q := aStar()
		for i:=0; i<1000000; i++ {
			q.Add(mknode(r.Float64()))
			if i % 3 == 0 {
				q.Take()
			}
		}
	}
}