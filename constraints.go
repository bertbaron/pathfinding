package solve

import "strconv"

// Constraint is a marker interface for constraints. Because the constraint methods refer to internal data structures
// we can not expose those methods (or can we somehow?)
type Constraint interface{}

// A possibly mutable constraint, returns true if a node is constraint, so it should not be expanded further.
type iconstraint interface {
	onVisit(node *node) bool
	onExpand(node *node) bool
	reset()
}

// value is irrelevant
type noConstraint bool

func (c noConstraint) onVisit(node *node) bool {
	return false
}

func (c noConstraint) onExpand(node *node) bool {
	return false
}

func (c noConstraint) reset() {}

func (c noConstraint) String() string {
	return "NoConstraint"
}

type noLoopConstraint struct {
	samefn func(State, State) bool
	depth  int
}

func (c noLoopConstraint) onVisit(node *node) bool {
	return false
}

func (c noLoopConstraint) onExpand(node *node) bool {
	ancestor := node.parent
	for i := 0; i < c.depth; i++ {
		if ancestor == nil {
			return false
		}
		if c.samefn(node.state, ancestor.state) {
			return true
		}
		ancestor = ancestor.parent
	}
	return false
}

func (c noLoopConstraint) reset() {}

func (c noLoopConstraint) String() string {
	return "NoLoopConstraint(" + strconv.Itoa(c.depth) + ")"
}

// NoLoopConstraint drops a state when it is equal to any ancestor state.
//
// A depth of 1 will compare only with the parent state. A depth of 2 will compare with
// the parent state and its parent state, etc.
//
// The states are compared using the provided function. Note that symmetric states my
// be considered equal by this function to eliminate symmetric branches from the search tree.
//
// Performance is linear in the depth or the actual search depth, whichever is smaller
func NoLoopConstraint(depth int, samefn func(State, State) bool) Constraint {
	return noLoopConstraint{samefn, depth}
}

// NoConstraint returns a constraint that will not drop states from the search tree
func NoConstraint() Constraint {
	return noConstraint(false)
}

type CPNode struct {
	value   float64
	visited bool
}

type CPMap interface {
	Get(state State) (CPNode, bool)
	Put(state State, value CPNode)
	Clear()
}

type cheapestPathConstraint struct {
	m CPMap
}

func (c cheapestPathConstraint) onExpand(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value < current.value {
		c.m.Put(node.state, CPNode{node.value, false})
		return false
	}
	return true
}

func (c cheapestPathConstraint) onVisit(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value < current.value || node.value == current.value && !current.visited {
		c.m.Put(node.state, CPNode{node.value, true})
		return false
	}
	return true
}

func (c cheapestPathConstraint) reset() {
	c.m.Clear()
}

func (c cheapestPathConstraint) String() string {
	return "CheapestPathConstraint"
}

// CheapestPathConstraint will drop a state when a cheaper path was found to an equal state. If two equal states have the
// same cost, than any of those states will be dropped the lowest cost.
//
// A custom map implementation needs to be provided to efficiently store the state. Note that symmetric states may map
// to the same key to eliminate symmetric branches from the search tree.
//
// Performance is constant time, but memory usage is linear to the number of states. Therefore this constraint
// is most usable in combination with A* or Breadth-First.
func CheapestPathConstraint(m CPMap) Constraint {
	return cheapestPathConstraint{m}
}
