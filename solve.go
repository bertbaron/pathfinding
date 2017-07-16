// Package solve provides algorithms like A*, IDA* and Depth-first
package solve

import (
	"math"
)

// Context can be used to interact with the solver and to maintain a custom context
// during the search.
type Context struct {
	Custom interface{}
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
}

// Result of the search
type Result struct {
	// The list of states leading from the root state to the goal state. If no solution
	// is found this list will be empty
	Solution []State

	// Number of nodes visited (dequeued) by the algorithm
	Visited  int

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

	next     *func() result
}

// ubound is an underbound for goal nodes. This is needed when IDA* is used to find multiple goal nodes to skip previously generated goal nodes
// limit is the maximum cost limit (inclusive)
// contour should be set to math.Inf(1). This value is set to the lowest cost encouterd >limit. The parameter is needed when recursively invoked
func generalSearch(queue strategy, visited int, expanded int, constr iconstraint, ubound float64, limit float64, contour float64, context Context) result {
	for {
		n := queue.Take()
		if n == nil {
			return result{nil, contour, visited, expanded, nil}
		}
		visited++
		if constr.onVisit(n) {
			continue
		}
		if n.state.IsGoal(context) && n.value > ubound {
			next := func() result {
				return generalSearch(queue, visited, expanded, constr, ubound, limit, contour, context)
			}
			return result{n, contour, visited, expanded, &next}
		}
		for _, child := range n.state.Expand(context) {
			childNode := &node{n, child, math.Max(n.value, child.Cost(context) + child.Heuristic(context))}
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

func idaStar(rootState State, constraint iconstraint, contour float64, ubound float64, limit float64, context Context, nextfn *func() result) result {
	visited := 0
	expanded := 0
	for true {
		var lastResult result
		if nextfn != nil {
			fn := *nextfn
			nextfn = nil
			lastResult = fn()
		} else {
			s := depthFirst()
			s.Add(&node{nil, rootState, rootState.Cost(context) + rootState.Heuristic(context)})
			constraint.reset()
			lastResult = generalSearch(s, visited, expanded, constraint, ubound, contour, math.Inf(1), context)
		}
		if lastResult.node != nil {
			// Found a solution
			underlying := lastResult.next
			nextIdaStarFn := func() result {
				return idaStar(rootState, constraint, contour, ubound, limit, context, underlying)
			}
			lastResult.next = &nextIdaStarFn
			return lastResult
		}
		if lastResult.contour > limit || math.IsInf(lastResult.contour, 1) || math.IsNaN(lastResult.contour) {
			// No (more) solutions
			lastResult.next = nil
			return lastResult
		}
		lastResult.next = nil
		ubound = contour
		visited = lastResult.visited
		expanded = lastResult.expanded
		contour = lastResult.contour
	}
	panic("Shouldn't be reached")
}

func toSlice(node *node) []State {
	if node == nil {
		return make([]State, 0)
	}
	return append(toSlice(node.parent), node.state)
}

func toResult(r *result) Result {
	return Result{toSlice(r.node), r.visited, r.expanded}
}

type solver struct {
	rootState  State
	algorithm  Algorithm
	constraint Constraint
	limit      float64
	context    interface{}

	started    bool
	result     *result
}

func solve(ss *solver) Result {
	if ss.started {
		if ss.result.next == nil {
			// no more possible solutions
			return Result{[]State{}, ss.result.visited, ss.result.expanded}
		}
		nextResult := (*ss.result.next)()
		ss.result = &nextResult
		return toResult(ss.result)
	}
	ss.started = true
	context := Context{ss.context}
	constraint := ss.constraint.(iconstraint)
	if ss.algorithm == IDAstar {
		nextResult := idaStar(ss.rootState, constraint, 0.0, -1.0, ss.limit, context, nil)
		ss.result = &nextResult
		return toResult(ss.result)
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

	constraint.reset()
	nextResult := generalSearch(s, 0, 0, constraint, -1.0, ss.limit, math.Inf(1), context)
	ss.result = &nextResult
	return toResult(ss.result)
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
	Context(context interface{}) Solver

	// Solves the problem returning the result
	Solve() Result
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

func (s *solver) Context(context interface{}) Solver {
	s.context = context
	return s
}

func (s *solver) Solve() Result {
	return solve(s)
}

// NewSolver creates a new solver
func NewSolver(rootState State) Solver {
	return &solver{rootState, Astar, NoConstraint(), math.Inf(1), nil, false, nil}
}
