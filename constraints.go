package solve

import "strconv"

// Constraint is a marker interface for constraints. Because the constraint methods refer to internal data structures
// we can not expose those methods
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
// The states are compared using the provided function. Note that symmetric states may
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

// CPMap needs to be implemented to efficiently store the state of the CheapestPathConstraint.
//
// A typical implementation will look something like:
//
//  type cpMap map[int64]float64
//
//  func (c cpMap) Get(state solve.State) (float64, bool) {
//  	value, ok := c[state.(swapState).id]
//  	return value, ok
//  }
//
//  func (c cpMap) Put(state solve.State, value float64) {
//  	c[state.(swapState).id] = value
//  }
//
//  func (c *cpMap) Clear() {
//  	*c = make(cpMap)
//  }
type CPMap interface {
	Get(state State) (float64, bool)
	Put(state State, value float64)
	Clear()
}

type cheapestPathConstraint struct {
	m CPMap
}

func (c cheapestPathConstraint) onExpand(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value < current {
		c.m.Put(node.state, node.value)
		return false
	}
	return true
}

func (c cheapestPathConstraint) onVisit(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value <= current {
		c.m.Put(node.state, node.value)
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
// same cost, than any of those states will be dropped.
//
// A custom map implementation needs to be provided to efficiently store the state. Note that symmetric states may map
// to the same key to eliminate symmetric branches from the search tree.
//
// Performance is constant time, but memory usage is linear to the number of states. Therefore this constraint
// is most usable in combination with A* or Breadth-First.
func CheapestPathConstraint(m CPMap) Constraint {
	return cheapestPathConstraint{m}
}
