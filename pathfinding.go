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

type State interface {
	Expand() []State
	Cost() float64
	Heuristic() float64
	IsGoal() bool
}

type Node struct {
	parent *Node
	state  State
	value  float64
}

type strategy interface {
	Take() *Node
	Add(node *Node)
}

// A PriorityQueue implements heap.Interface and holds Nodes
type priorityQueue []*Node

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
	item := x.(*Node)
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
func (pq *priorityQueue) Take() *Node {
	if len(*pq) == 0 {
		return nil
	}
	return heap.Pop(pq).(*Node)
}

// strategy
func (pq *priorityQueue) Add(node *Node) {
	heap.Push(pq, node)
}

type lifo []*Node

func (dfq *lifo) Take() *Node {
	if len(*dfq) == 0 {
		return nil
	}
	old := *dfq
	n := len(old)
	item := old[n - 1]
	*dfq = old[0 : n - 1]
	return item
}

func (dfq *lifo) Add(node *Node) {
	*dfq = append(*dfq, node)
}

// A possibly mutable constraint, returns true if a node is constraint, so it should not be expanded further.
type constraint interface {
	onVisit(node *Node) bool
	onExpand(node *Node) bool
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

func (c noConstraint) onVisit(node *Node) bool {
	return false
}

func (c noConstraint) onExpand(node *Node) bool {
	return false
}

type result struct {
	node     *Node
	contour  float64
	visited  int
	expanded int
}

func generalSearch(queue strategy, visited int, expanded int, constr constraint, limit float64) result {
	contour := math.MaxFloat64

	for {
		node := queue.Take()
		if node == nil {
			return result{nil, contour, visited, expanded}
		}
		visited++
		if constr.onVisit(node) {
			continue
		}
		if node.state.IsGoal() {
			return result{node, contour, visited, expanded}
		}
		for _, child := range node.state.Expand() {
			childNode := &Node{node, child, math.Max(node.value, child.Cost() + child.Heuristic())}
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
		s.Add(&Node{nil, rootState, rootState.Cost() + rootState.Heuristic()})
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

type Result struct {
	Solution *Node
	Visited  int
	Expanded int
}

func Solve(rootState State, algorithm Algorithm, limit float64) Result {
	if algorithm == IDAstar {
		result := idaStar(rootState, limit)
		return Result{result.node, result.visited, result.expanded}
	}
	var s strategy
	switch algorithm {
	case Astar: s = aStar()
	case DepthFirst: s = depthFirst()
	}
	s.Add(&Node{nil, rootState, rootState.Cost() + rootState.Heuristic()})
	var constraint noConstraint

	result := generalSearch(s, 0, 0, constraint, limit)
	return Result{result.node, result.visited, result.expanded}
}