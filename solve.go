package solve

import (
	"math"
)

type State interface {
	Cost() float64
	IsGoal() bool
	Expand() []State
	Heuristic() float64
	Id() interface{}
}

type Node interface {
	Parent() Node
	State() State
	Exists() bool
}

type Result struct {
	Solution Node
	Visited  int
	Expanded int
}

type node struct {
	parent *node
	state  State
	value  float64
}

func (node node) Parent() Node {
	return node.parent
}

func (node node) State() State {
	return node.state
}

func (node *node) Exists() bool {
	return node != nil
}

type result struct {
	node     *node
	contour  float64
	visited  int
	expanded int
}

func generalSearch(queue strategy, visited int, expanded int, constr iconstraint, limit float64) result {
	contour := math.MaxFloat64

	for {
		n := queue.Take()
		if n == nil {
			return result{nil, contour, visited, expanded}
		}
		visited++
		if constr.onVisit(n) {
			continue
		}
		if n.state.IsGoal() {
			return result{n, contour, visited, expanded}
		}
		for _, child := range n.state.Expand() {
			childNode := &node{n, child, math.Max(n.value, child.Cost() + child.Heuristic())}
			if constr.onExpand(childNode) {
				continue
			}
			if childNode.value > limit {
				contour = math.Min(contour, childNode.value)
				continue
			}
			queue.Add(childNode)
			expanded++

		}

	}
	return result{nil, contour, visited, expanded}
}

func idaStar(rootState State, constraint Constraint, limit float64) result {
	visited := 0
	expanded := 0
	contour := 0.0
	for true {
		// TODO doesn't stop when contour is infinity when there is no limit?
		s := depthFirst()
		s.Add(&node{nil, rootState, rootState.Cost() + rootState.Heuristic()})
		lastResult := generalSearch(s, visited, expanded, createConstraint(constraint), contour)
		if lastResult.node != nil || lastResult.contour > limit {
			return lastResult
		}
		visited = lastResult.visited
		expanded = lastResult.expanded
		contour = lastResult.contour
	}
	panic("Shouldn't be reached")
}

func solve(ss solver) Result {
	if ss.algorithm == IDAstar {
		result := idaStar(ss.rootState, ss.constraint, ss.limit)
		return Result{result.node, result.visited, result.expanded}
	}
	var s strategy
	switch ss.algorithm {
	case Astar: s = aStar()
	case DepthFirst: s = depthFirst()
	}
	s.Add(&node{nil, ss.rootState, ss.rootState.Cost() + ss.rootState.Heuristic()})

	result := generalSearch(s, 0, 0, createConstraint(ss.constraint), ss.limit)
	return Result{result.node, result.visited, result.expanded}
}

type Solver interface {
	Algorithm(algorithm Algorithm) Solver
	Constraint(constraint Constraint) Solver
	Limit(limit float64) Solver
	Solve() Result
}
type solver struct {
	rootState  State
	algorithm  Algorithm
	constraint Constraint
	limit      float64
}

func (s *solver) Algorithm(algorithm Algorithm) Solver {
	s.algorithm = algorithm
	return s
}

func (s *solver) Constraint(constraint Constraint) Solver {
	s.constraint = constraint
	return s
}

func (s *solver) Limit(limit float64) Solver {
	s.limit = limit
	return s
}

func (s *solver) Solve() Result {
	return solve(*s)
}

func NewSolver(rootState State) Solver {
	return &solver{rootState, Astar, NONE, math.MaxFloat64}
}
