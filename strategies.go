package solve

import "container/heap"

type Algorithm int

const (
	Astar Algorithm = iota
	DepthFirst Algorithm = iota
	BreadthFirst Algorithm = iota
	IDAstar Algorithm = iota
)

func (a Algorithm) String() string {
	switch a {
	case Astar: return "A*"
	case IDAstar: return "IDA*"
	case BreadthFirst: return "BreadthFirst"
	case DepthFirst: return "DepthFirst"
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
	item := old[n - 1]
	*pq = old[0 : n - 1]
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
	item := old[n - 1]
	*dfq = old[0 : n - 1]
	return item
}

func (dfq *lifo) Add(node *node) {
	*dfq = append(*dfq, node)
}

// Breadth-first strategy, based on a fifo queue
// We might replace this with a ring-buffer for performance
type fifo []*node

func (bfq *fifo) Take() *node {
	if len(*bfq) == 0 {
		return nil
	}
	old := *bfq
	n := len(old)
	item := old[0]
	*bfq = old[1 : n]
	return item
}

func (bfq *fifo) Add(node *node) {
	*bfq = append(*bfq, node)
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
	queue := make(fifo, 0, 64)
	return &queue
}
