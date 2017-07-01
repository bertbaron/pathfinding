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
