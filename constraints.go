package pathfinding

import (
	"math"
	"fmt"
	"container/heap"
)

type Constraint int

const (
	NONE Constraint = iota
	NO_RETURN Constraint = iota
	NO_LOOP Constraint = iota
	CHEAPEST_PATH Constraint = iota
)

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

// value is irrelevant
type noConstraint bool

func (c noConstraint) onExpand(node *node) bool {
	return false
}

func (c noConstraint) onVisit(node *node) bool {
	return false
}


// value is the limit for looking back
type noLoopConstraint int

func (c noLoopConstraint) onExpand(node *node) bool {
	return false
}

func (c noLoopConstraint) onVisit(node *node) bool {
	id := node.State().Id()
	ancestor := node.parent
	for i := 0; i < int(c); i++ {
		if ancestor == nil {
			return false
		}
		if (id == ancestor.state.Id()) {
			return true
		}
		ancestor = ancestor.parent
	}
	return false
}

type constraintNode struct {
	value   float64
	visited bool
}

type cheapestPathConstraint map[interface{}]constraintNode

func (c cheapestPathConstraint) onExpand(node *node) bool {
	id := node.state.Id()
	current, ok := c[id]
	if !ok || node.value < current.value {
		c[id] = constraintNode{node.value, false}
		return false
	}
	return true
}

func (c cheapestPathConstraint) onVisit(node *node) bool {
	id := node.state.Id()
	current, ok := c[id]
	if !ok || node.value < current.value || node.value == current.value && !current.visited {
		c[id] = constraintNode{node.value, true}
		return false
	}
	return true
}

func createConstraint(constraint Constraint) iconstraint {
	switch constraint {
	case NONE: return noConstraint(false)
	case NO_RETURN: return noLoopConstraint(2)
	case NO_LOOP: return noLoopConstraint(math.MaxInt32)
	case CHEAPEST_PATH: return make(cheapestPathConstraint)
	}
	panic(fmt.Sprintf("invalid constraint: %v", constraint))
}
