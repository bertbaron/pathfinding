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
	Visited int

	// Number of nodes expanded (enqueued) by the algorithm
	Expanded int
}

// Solved returns true if the result yields a solution
func (r Result) Solved() bool {
	return len(r.Solution) > 0
}

// GoalState returns the last state of Solution(). Can only be called if r.Solved() == true
func (r Result) GoalState() State {
	return r.Solution[len(r.Solution)-1]
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

	next *func() result
}

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

func startGeneralSearch(queue strategy, constr iconstraint, limit float64, context Context) result {
	return generalSearch(queue, 0, 0, constr, -1.0, limit, math.Inf(1), context)
}

func idaStar(rootState State, constraint iconstraint, contour float64, ubound float64, limit float64, context Context, nextfn *func() result) result {
	visited := 0
	expanded := 0
	for true {
		var lastResult result
		if nextfn == nil {
			// start with new iteration
			s := depthFirst()
			s.Add(&node{nil, rootState, rootState.Cost(context) + rootState.Heuristic(context)})
			constraint.reset()
			lastResult = generalSearch(s, visited, expanded, constraint, ubound, contour, math.Inf(1), context)
		} else {
			// continue previous iteration
			fn := *nextfn
			nextfn = nil
			lastResult = fn()
		}
		if lastResult.node != nil {
			// Found a solution
			underlyingNextFn := lastResult.next
			nextIdaStarFn := func() result {
				return idaStar(rootState, constraint, contour, ubound, limit, context, underlyingNextFn)
			}
			lastResult.next = &nextIdaStarFn
			return lastResult
		}
		lastResult.next = nil
		if lastResult.contour > limit || math.IsInf(lastResult.contour, 1) || math.IsNaN(lastResult.contour) {
			// no (more) solutions
			return lastResult
		}
		ubound = contour
		visited = lastResult.visited
		expanded = lastResult.expanded
		contour = lastResult.contour
	}
	panic("Shouldn't be reached")
}

func startIdaStar(rootState State, constraint iconstraint, limit float64, context Context) result {
	return idaStar(rootState, constraint, 0.0, -1.0, limit, context, nil)
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

	started bool
	result  *result
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
		nextResult := startIdaStar(ss.rootState, constraint, ss.limit, context)
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
	nextResult := startGeneralSearch(s, constraint, ss.limit, context)
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

	// Convenience method for finding all solutions. The solutions are put in the
	// given channel. The channel is closed when the solver is completed.
	SolveAll(solutions chan<- Result)

	// True if the search is completed
	Completed() bool
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

func (s *solver) SolveAll(solutions chan<- Result) {
	defer close(solutions)
	for result := s.Solve(); result.Solved(); result = s.Solve() {
		solutions <- result
	}
}

func (s *solver) Completed() bool {
	return s.started && s.result.next == nil
}

// NewSolver creates a new solver
func NewSolver(rootState State) Solver {
	return &solver{rootState, Astar, NoConstraint(), math.Inf(1), nil, false, nil}
}
