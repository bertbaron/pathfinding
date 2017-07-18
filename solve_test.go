package solve

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"unicode"
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

func (s state) Cost(ctx Context) float64 {
	return s.cost
}

func (s state) IsGoal(ctx Context) bool {
	return unicode.IsUpper([]rune(s.node)[0])
}

func (s state) Expand(ctx Context) []State {
	var children []State
	if edges, ok := s.graph[s.node]; ok {
		for _, edge := range edges {
			children = append(children, expand(s, edge))
		}
	}
	return children
}

func (s state) Heuristic(ctx Context) float64 {
	return 0
}

// for no-loop-constraint
func same(a, b State) bool {
	return a.(state).node == b.(state).node
}

// for cheapest-path-constraint
type cpMap map[string]CPNode

func (c cpMap) Get(s State) (CPNode, bool) {
	value, ok := c[s.(state).node]
	return value, ok
}

func (c cpMap) Put(s State, value CPNode) {
	c[s.(state).node] = value
}
func (c *cpMap) Clear() {
	*c = make(cpMap)
}

var testNoConstraint = NoConstraint()
var testNoReturnConstraint = NoLoopConstraint(2, same)
var testNoLoopConstraint = NoLoopConstraint(99999, same)
var testCPMap = make(cpMap)
var testCheapestPathConstraint = CheapestPathConstraint(&testCPMap)

type goalCost struct {
	goal string
	cost float64
}

func equalGoalCost(a, b []goalCost) bool {
	if len(a) != len(b) {
		return false
	}
	for i, value := range a {
		if b[i] != value {
			return false
		}
	}
	return true
}

type sortableGoals []goalCost

func (s sortableGoals) Len() int {
	return len(s)
}

func (s sortableGoals) Less(i, j int) bool {
	return s[i].goal < s[j].goal
}

func (s sortableGoals) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func solveAll(solver Solver) []goalCost {
	results := make([]goalCost, 0)
	for result := solver.Solve(); result.Solved(); result = solver.Solve() {
        goalState := result.Solution[len(result.Solution)-1].(state)
        results = append(results, goalCost{goalState.node, goalState.cost})
	}
	return results
}

func testSolve(t *testing.T, graph graph, algorithm Algorithm, constraint Constraint, limit float64, expected []goalCost) {
	solver := NewSolver(create(graph)).
		Algorithm(algorithm).
		Constraint(constraint).
		Limit(limit)
	actual := solveAll(solver)

	name := fmt.Sprintf("(%v,%v)", algorithm, constraint)
	if algorithm == Astar || algorithm == BreadthFirst || algorithm == IDAstar {
		if !equalGoalCost(actual, expected) {
			t.Errorf("%v - Expected %v but found %v", name, expected, actual)
		}
		return
	}

	if algorithm == DepthFirst {
		sort.Sort(sortableGoals(expected))
		sort.Sort(sortableGoals(actual))
		if !equalGoalCost(actual, expected) {
			t.Errorf("%v - Expected %v but found %v", name, expected, actual)
		}
		return
	}
}

// Solves the problem with all algorithms and constraints that should return in the optimal solution
func testSolveAllAlgorithms(t *testing.T, graph graph, includeBF bool, expected []goalCost) {
	testSolve(t, graph, Astar, testNoConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, Astar, testNoReturnConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, Astar, testNoLoopConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, Astar, testCheapestPathConstraint, math.MaxFloat64, expected)

	testSolve(t, graph, IDAstar, testNoConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, IDAstar, testNoReturnConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, IDAstar, testNoLoopConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, IDAstar, testCheapestPathConstraint, math.MaxFloat64, expected)

	testSolve(t, graph, DepthFirst, testNoConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, DepthFirst, testNoReturnConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, DepthFirst, testNoLoopConstraint, math.MaxFloat64, expected)
	testSolve(t, graph, DepthFirst, testCheapestPathConstraint, math.MaxFloat64, expected)

	// BF is only optimal if the length of costs corresonds with the length of the path
	if includeBF {
		testSolve(t, graph, BreadthFirst, testNoConstraint, math.MaxFloat64, expected)
		testSolve(t, graph, BreadthFirst, testNoReturnConstraint, math.MaxFloat64, expected)
		testSolve(t, graph, BreadthFirst, testNoLoopConstraint, math.MaxFloat64, expected)
		testSolve(t, graph, BreadthFirst, testCheapestPathConstraint, math.MaxFloat64, expected)
	}
}

func TestSimpleProblem(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"b", 1}, {"c", 1}}
	g["b"] = []edge{{"D", 1}, {"c", 1}}
	expected := []goalCost{{"D", 2}}
	testSolveAllAlgorithms(t, g, true, expected)
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
	expected := []goalCost{{"D", 21}, {"C", 116.0}, {"B", 202.0}}
	testSolveAllAlgorithms(t, g, false, expected)
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
	if len(result.Solution) != 1 {
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

func TestStatisticsWithDifferentConstraints(t *testing.T) {
	g := make(graph)
	g["a"] = []edge{{"a", 1}, {"b", 1}}
	g["b"] = []edge{{"c", 1}, {"d", 2}}
	g["c"] = []edge{{"a", 1}, {"d", 1}}
	g["d"] = []edge{{"E", 1}}
	testStatistics(t, g, Astar, testNoConstraint, 27, 16) // currently the test depends on the queue implementation...
	testStatistics(t, g, Astar, testNoReturnConstraint, 8, 7)
	testStatistics(t, g, Astar, testNoLoopConstraint, 6, 6)
	testStatistics(t, g, Astar, testCheapestPathConstraint, 4, 5)
}

type dummyState struct {
	State
	name string
}

func dummyNode(parent *node, name string, costs float64) *node {
	return &node{parent, dummyState{nil, name}, costs}
}

// for no-loop-constraint
func equalDummyStates(a, b State) bool {
	return a.(dummyState).name == b.(dummyState).name
}

func TestNoLoopConstraint(t *testing.T) {
	assert := func(name string, value, expected interface{}) {
		if value != expected {
			t.Errorf("%v - Expected %v, but was %v", name, expected, value)
		}
	}

	c := NoLoopConstraint(2, equalDummyStates).(iconstraint)
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
	for i := 0; i < 1000; i++ {
		b.Add(mknode(i))
		if i%3 == 0 {
			taken := b.Take()
			if taken == nil {
				t.Errorf("Expected node %v at head of the buffer, but the buffer was empty", lastTaken+1)
				return
			}
			if int(taken.value) != lastTaken+1 {
				t.Errorf("Expected element %v from the buffer, but was %v", lastTaken+1, taken.value)
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
		for i := 0; i < 3000000; i++ {
			b.Add(node)
			if i%3 == 0 {
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
		for i := 0; i < 1000000; i++ {
			q.Add(mknode(r.Float64()))
			if i%3 == 0 {
				q.Take()
			}
		}
	}
}

func BenchmarkAStarStrategyDiscrete(b *testing.B) {
	// for Astar we can not reuse the node, so this test involves more overhead
	mknode := func(value float64) *node {
		return &node{nil, nil, value}
	}

	r := rand.New(rand.NewSource(123))
	for n := 0; n < b.N; n++ {
		q := aStar()
		for i := 0; i < 1000000; i++ {
			q.Add(mknode(float64(r.Intn(100))))
			if i%3 == 0 {
				q.Take()
			}
		}
	}
}
