package solve

import (
	"fmt"
	"math"
)

// Constraint that can be used to eliminate nodes from the search tree
type Constraint int

const (
	// NO_CONSTRAINT will not drop states from the search tree
	NO_CONSTRAINT Constraint = iota

	// NO_RETURN drops state when it is equal to its parent or grandparent
	//
	// Performance is constant time, almost always pays off when applicable
	NO_RETURN Constraint = iota

	// NO_LOOP drops a state when it is equal to any ancestor state. NO_LOOP is more generic than NO_RETURN.
	//
	// Performance is linear to the search depth
	NO_LOOP Constraint = iota

	// CHEAPEST_PATH will drop a state when a cheaper path was found to an equal state. If two equal states have the
	// same cost, than any of those states will be dropped.
	// the lowest cost. CHEAPEST_PATH is more more generic than NO_LOOP
	//
	// Performance is constant time, but memory usage is linear to the number of states. Therefore this constraint
	// is most usable in combination with A* or Breadth-First.
	CHEAPEST_PATH Constraint = iota
)

func (c Constraint) String() string {
	switch c {
	case NO_CONSTRAINT:
		return "None"
	case NO_RETURN:
		return "No_return"
	case NO_LOOP:
		return "No_loop"
	case CHEAPEST_PATH:
		return "Cheapest_path"
	}
	return "<unknown>"
}

// A possibly mutable constraint, returns true if a node is constraint, so it should not be expanded further.
type iconstraint interface {
	onVisit(node *node) bool
	onExpand(node *node) bool
}

// value is irrelevant
type noConstraint bool

func (c noConstraint) onVisit(node *node) bool {
	return false
}

func (c noConstraint) onExpand(node *node) bool {
	return false
}

// value is the limit for looking back
type noLoopConstraint int

func (c noLoopConstraint) onVisit(node *node) bool {
	return false
}

func (c noLoopConstraint) onExpand(node *node) bool {
	id := node.state.Id()
	ancestor := node.parent
	for i := 0; i < int(c); i++ {
		if ancestor == nil {
			return false
		}
		if id == ancestor.state.Id() {
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
	case NO_CONSTRAINT:
		return noConstraint(false)
	case NO_RETURN:
		return noLoopConstraint(2)
	case NO_LOOP:
		return noLoopConstraint(math.MaxInt32)
	case CHEAPEST_PATH:
		return make(cheapestPathConstraint)
	}
	panic(fmt.Sprintf("invalid constraint: %v", constraint))
}
