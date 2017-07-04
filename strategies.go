package solve

import (
	"container/heap"
	"fmt"
	"strings"
	"sort"
)

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

func compare(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func (pq priorityQueue) Less(i, j int) bool {
	diff := compare(pq[i].value, pq[j].value)
	if diff == 0 {
		diff = compare(pq[j].state.Cost(), pq[i].state.Cost())
	}
	return diff < 0
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

func (pq priorityQueue) String() string {
	copy := pq
	sort.Sort(copy)
	var ss []string
	for _, item := range copy {
		ss = append(ss, fmt.Sprintf("%v (%v)", item.state, item.value))
	}
	return strings.Join(ss, ", ")
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

// Inspired by github.com/phf/go-queue/queue, but we implement our own
// because we only need part of the functionality and can make a sligtly
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
