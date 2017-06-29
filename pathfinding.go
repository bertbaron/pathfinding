package main

import (
	"container/heap"
	"math"
	"fmt"
	"time"
)

type Algorithm int

const (
	Astar Algorithm = iota
	DepthFirst Algorithm = iota
	BreadthFirst Algorithm = iota
	IDAstar Algorithm = iota
)

type Constraint int

const (
	NONE Constraint = iota
	NO_RETURN Constraint = iota
	NO_LOOP Constraint = iota
	SHORTEST_PATH_TO_STATE Constraint = iota
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

type strategy interface {
	Take() *node
	Add(node *node)
}

// A PriorityQueue implements heap.Interface and holds Nodes
type priorityQueue []*node

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].value < pq[j].value
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*node)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n - 1]
	*pq = old[0 : n - 1]
	return item
}

// strategy
func (pq *priorityQueue) Take() *node {
	if len(*pq) == 0 {
		return nil
	}
	return heap.Pop(pq).(*node)
}

// strategy
func (pq *priorityQueue) Add(node *node) {
	heap.Push(pq, node)
}

type lifo []*node

func (dfq *lifo) Take() *node {
	if len(*dfq) == 0 {
		return nil
	}
	old := *dfq
	n := len(old)
	item := old[n - 1]
	*dfq = old[0 : n - 1]
	return item
}

func (dfq *lifo) Add(node *node) {
	*dfq = append(*dfq, node)
}

// A possibly mutable constraint, returns true if a node is constraint, so it should not be expanded further.
type iconstraint interface {
	onVisit(node *node) bool
	onExpand(node *node) bool
}

func aStar() strategy {
	pq := make(priorityQueue, 0, 64)
	heap.Init(&pq)
	return &pq
}

func depthFirst() strategy {
	queue := make(lifo, 0, 64)
	return &queue
}

type noConstraint bool

func (c noConstraint) onVisit(node *node) bool {
	return false
}

func (c noConstraint) onExpand(node *node) bool {
	return false
}

type noReturnConstraint bool

func (c noReturnConstraint) onVisit(node *node) bool {
	fmt.Println("test")
	return node.parent != nil && node.parent.parent != nil && node.parent.parent.state.Id() == node.state.Id()
}

func (c noReturnConstraint) onExpand(node *node) bool {
	return false
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

func idaStar(rootState State, limit float64) result {
	visited := 0
	expanded := 0
	contour := 0.0
	for true {
		start := time.Now()
		// TODO doesn't stop when contour is infinity when there is no limit?
		s := depthFirst()
		s.Add(&node{nil, rootState, rootState.Cost() + rootState.Heuristic()})
		var constraint noConstraint
		lastResult := generalSearch(s, visited, expanded, constraint, contour)
		if lastResult.node != nil || lastResult.contour > limit {
			return lastResult
		}

		seconds := float64(time.Since(start).Seconds())
		rate := float64(lastResult.visited - visited) / seconds
		fmt.Printf("Contour: %f, expanded: %d, %.0f/s\n", contour, visited, rate)
		visited = lastResult.visited
		expanded = lastResult.expanded
		contour = lastResult.contour
	}
	panic("Shouldn't be reached")
}

func Solve(rootState State, algorithm Algorithm, constraint Constraint, limit float64) Result {
	if algorithm == IDAstar {
		result := idaStar(rootState, limit)
		return Result{result.node, result.visited, result.expanded}
	}
	var s strategy
	switch algorithm {
	case Astar: s = aStar()
	case DepthFirst: s = depthFirst()
	}
	s.Add(&node{nil, rootState, rootState.Cost() + rootState.Heuristic()})

	var constr iconstraint
	switch constraint {
	case NONE: constr = noConstraint(false)
	case NO_RETURN: constr = noReturnConstraint(false)
	}

	result := generalSearch(s, 0, 0, constr, limit)
	return Result{result.node, result.visited, result.expanded}
}