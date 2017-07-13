package solve

import (
	"container/heap"
)

// Algorithm to be used to solve the problem
type Algorithm int

const (
	// Astar (A*) solves the problem in the least number of visited nodes.
	// Requires a lot of memory.
	//
	// Wil return the optimal solution if the heuristic is admissible
	Astar Algorithm = iota

	// DepthFirst uses backtracking to traverse the search tree.
	// Requires very little memory.
	//
	// This is ideal for problems that simply require a solution but don't have some notion of 'best' solution
	// (like a Sudoku puzzle).
	//
	// Will not guarantee to find the optimal solution
	DepthFirst Algorithm = iota

	// BreadthFirst expands all nodes at a specific depth before going to the next depth.
	// Requires a lot of memory.
	//
	// It is almost always better to use A* or IDA*. If the cost is proportional to the depth however and no
	// heuristic is provided then breadth-first is equivalent to A* but faster.
	//
	// Will find the optimal solution if the shortest path is the optimal solution
	BreadthFirst Algorithm = iota

	// IDAstar (IDA*, Iterative Deepening A*) performs iterative depth-first searches to find the optimal solution
	// while using very little memory. This works very well when the costs increase in discrete steps along the
	// path. Because each iteration repeats the work from the previous iteration however, many more nodes may be
	// visited to find the solution than will be the case with A*
	//
	// Will find the optimal solution if the heuristic is admissible
	IDAstar Algorithm = iota
)

func (a Algorithm) String() string {
	switch a {
	case Astar:
		return "A*"
	case IDAstar:
		return "IDA*"
	case BreadthFirst:
		return "BreadthFirst"
	case DepthFirst:
		return "DepthFirst"
	}
	return "<unknown>"
}

type strategy interface {
	Take() *node
	Add(node *node)
}

// A* strategy, based on a priority queue
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
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) Take() *node {
	if len(*pq) == 0 {
		return nil
	}
	return heap.Pop(pq).(*node)
}

func (pq *priorityQueue) Add(node *node) {
	heap.Push(pq, node)
}

// Depth-first strategy, based on a lifo queue
type lifo []*node

func (dfq *lifo) Take() *node {
	if len(*dfq) == 0 {
		return nil
	}
	old := *dfq
	n := len(old)
	item := old[n-1]
	*dfq = old[0 : n-1]
	return item
}

func (dfq *lifo) Add(node *node) {
	*dfq = append(*dfq, node)
}

// Inspired by github.com/phf/go-queue/queue, but we implement our own
// because we only need part of the functionality and can make a slightly
// more efficient implementation without the need for an external dependency
type ringbuffer struct {
	buffer []*node
	start  int
	end    int
}

func inc(b *ringbuffer, i int) int {
	return (i + 1) & (len(b.buffer) - 1) // requires l = 2^n
}

func (b *ringbuffer) Take() *node {
	if b.start == b.end {
		return nil
	}
	item := b.buffer[b.start]
	b.start = inc(b, b.start)
	return item
}

func (b *ringbuffer) Add(node *node) {
	b.buffer[b.end] = node
	b.end = inc(b, b.end)
	if b.start == b.end {
		grow(b)
	}
}

func grow(b *ringbuffer) {
	oldsize := len(b.buffer)
	size := oldsize * 2
	adjusted := make([]*node, size)
	if b.start < b.end {
		// not "wrapped" around, one copy suffices
		copy(adjusted, b.buffer[b.start:b.end])
	} else {
		// "wrapped" around, need two copies
		n := copy(adjusted, b.buffer[b.start:])
		copy(adjusted[n:], b.buffer[:b.end])
	}
	b.buffer = adjusted
	b.start = 0
	b.end = oldsize
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

func breadthFirst() strategy {
	var b ringbuffer
	b.buffer = make([]*node, 64)
	return &b
}
