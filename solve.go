// Package solve provides algorithms like A*, IDA* and Depth-first
package solve

import (
	"math"
)

// Context can be used to interact with the solver and to maintain a custom context
// during the search.
type Context struct {
	custom *interface{}
}
// The State representing a state in the search tree
//
// An implementation of this interface represents the problem. It tells the algorithm how
// to get from one state to another, how much it costs to reach the state etc.
type State interface {
	// The costs to reach this state
	Cost(ctx Context) float64

	// Returns true if this is a goal state
	IsGoal(ctx Context) bool

	// Expands this state in zero or more child states
	Expand(ctx Context) []State

	// Estimated costs to reach a goal. Use 0 for no heuristic. Most algorithms will
	// find the optimal solution if the heuristic is admissible, meaning it will never
	// over-estimate the costs to reach a goal
	Heuristic(ctx Context) float64

	// Returns an id that is used in constraints to reduce the search tree by
	// eliminating identical states. Can be nil if no constraint is used.
	//
	// Symmetrical states may return the same Id to eliminate even more branches of the
	// search tree
	//
	// The result must be comparable (like go map keys)
	Id() interface{}
}

// Result of the search
type Result struct {
	// The list of states leading from the root state to the goal state. If no solution
	// is found this list will be empty
	Solution []State

	// Number of nodes visited (dequeued) by the algorithm
	Visited int

	// Number of nodes expanded (enqueued) by the algorithm
	Expanded int
}

type node struct {
	parent *node
	state  State
	value  float64
}

type result struct {
	node     *node
	contour  float64
	visited  int
	expanded int
}

func generalSearch(queue strategy, visited int, expanded int, constr iconstraint, limit float64, context Context) result {
	contour := math.Inf(1)

	for {
		n := queue.Take()
		if n == nil {
			return result{nil, contour, visited, expanded}
		}
		visited++
		if constr.onVisit(n) {
			continue
		}
		if n.state.IsGoal(context) {
			return result{n, contour, visited, expanded}
		}
		for _, child := range n.state.Expand(context) {
			childNode := &node{n, child, math.Max(n.value, child.Cost(context)+child.Heuristic(context))}
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
}

func idaStar(rootState State, constraint Constraint, limit float64, context Context) result {
	visited := 0
	expanded := 0
	contour := 0.0
	for true {
		s := depthFirst()
		s.Add(&node{nil, rootState, rootState.Cost(context) + rootState.Heuristic(context)})
		lastResult := generalSearch(s, visited, expanded, createConstraint(constraint), contour, context)
		if lastResult.node != nil || lastResult.contour > limit || math.IsInf(lastResult.contour, 1) || math.IsNaN(lastResult.contour) {
			return lastResult
		}
		visited = lastResult.visited
		expanded = lastResult.expanded
		contour = lastResult.contour
		//fmt.Printf("contour: %v, visited: %v\n", contour, visited)
	}
	panic("Shouldn't be reached")
}

func toSlice(node *node) []State {
	if node == nil {
		return make([]State, 0)
	}
	return append(toSlice(node.parent), node.state)
}

func solve(ss solver) Result {
	context := Context{ss.context}
	if ss.algorithm == IDAstar {
		result := idaStar(ss.rootState, ss.constraint, ss.limit, context)
		return Result{toSlice(result.node), result.visited, result.expanded}
	}
	var s strategy
	switch ss.algorithm {
	case Astar:
		s = aStar()
	case DepthFirst:
		s = depthFirst()
	case BreadthFirst:
		s = breadthFirst()
	}
	s.Add(&node{nil, ss.rootState, ss.rootState.Cost(context) + ss.rootState.Heuristic(context)})

	result := generalSearch(s, 0, 0, createConstraint(ss.constraint), ss.limit, context)
	return Result{toSlice(result.node), result.visited, result.expanded}
}

// Solver to solve the problem.
type Solver interface {
	// The algorithm to use, defaults to IDAstar
	Algorithm(algorithm Algorithm) Solver

	// The constraint to use, defaults to NONE
	Constraint(constraint Constraint) Solver

	// The limit to use. The problem will not be exanded beyond this limit. Defaults
	// to math.Inf(1).
	Limit(limit float64) Solver

	// Custom context which is passed to the methods of the state. Can contain for example precalculated data that
	// is used to speed up calculations. Be careful with state in the context though.
	Context(context *interface{}) Solver

	// Solves the problem returning the result
	Solve() Result
}
type solver struct {
	rootState  State
	algorithm  Algorithm
	constraint Constraint
	limit      float64
	context    *interface{}
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

func (s *solver) Context(context *interface{}) Solver {
	s.context = context
	return s
}

func (s *solver) Solve() Result {
	return solve(*s)
}

// NewSolver creates a new solver
func NewSolver(rootState State) Solver {
	return &solver{rootState, Astar, NO_CONSTRAINT, math.Inf(1), nil}
}
