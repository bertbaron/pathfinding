package solve

//// Constraint that can be used to eliminate nodes from the search tree
//type Constraint int
//
//const (
//	// NO_CONSTRAINT will not drop states from the search tree
//	NO_CONSTRAINT Constraint = iota
//
//	// NO_RETURN drops state when it is equal to its parent or grandparent
//	//
//	// Performance is constant time, almost always pays off when applicable
//	NO_RETURN Constraint = iota
//
//	// NO_LOOP drops a state when it is equal to any ancestor state. NO_LOOP is more generic than NO_RETURN.
//	//
//	// Performance is linear to the search depth
//	NO_LOOP Constraint = iota
//
//	// CHEAPEST_PATH will drop a state when a cheaper path was found to an equal state. If two equal states have the
//	// same cost, than any of those states will be dropped.
//	// the lowest cost. CHEAPEST_PATH is more more generic than NO_LOOP
//	//
//	// Performance is constant time, but memory usage is linear to the number of states. Therefore this constraint
//	// is most usable in combination with A* or Breadth-First.
//	CHEAPEST_PATH Constraint = iota
//)

//func (c Constraint) String() string {
//	switch c {
//	case NO_CONSTRAINT:
//		return "None"
//	case NO_RETURN:
//		return "No_return"
//	case NO_LOOP:
//		return "No_loop"
//	case CHEAPEST_PATH:
//		return "Cheapest_path"
//	}
//	return "<unknown>"
//}

// Constraint is a marker interface for constraints. Because the constraint methods refer to internal data structures
// we can not expose those methods (or can we somehow?)
type Constraint interface {

}

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

func (c noConstraint) reset() {

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

func (c noLoopConstraint) reset() {

}

type constraintNode struct {
	value   float64
	visited bool
}

type cheapestPathConstraint struct {
	m map[interface{}]constraintNode
}

func (c *cheapestPathConstraint) onExpand(node *node) bool {
	id := node.state.Id()
	current, ok := c.m[id]
	if !ok || node.value < current.value {
		c.m[id] = constraintNode{node.value, false}
		return false
	}
	return true
}

func (c *cheapestPathConstraint) onVisit(node *node) bool {
	id := node.state.Id()
	current, ok := c.m[id]
	if !ok || node.value < current.value || node.value == current.value && !current.visited {
		c.m[id] = constraintNode{node.value, true}
		return false
	}
	return true
}

func (c *cheapestPathConstraint) reset() {
	c.m = make(map[interface{}]constraintNode)
}

func NoConstraint() Constraint {
	return noConstraint(false)
}

func CheapestPathConstraint() Constraint {
	return &cheapestPathConstraint{make(map[interface{}]constraintNode)}
}

//func createConstraint(constraint Constraint) iconstraint {
//	switch constraint {
//	case NO_CONSTRAINT:
//		return noConstraint(false)
//	case NO_RETURN:
//		return noLoopConstraint(2)
//	case NO_LOOP:
//		return noLoopConstraint(math.MaxInt32)
//	case CHEAPEST_PATH:
//		return &cheapestPathConstraint{make(map[interface{}]constraintNode)}
//	}
//	panic(fmt.Sprintf("invalid constraint: %v", constraint))
//}

type ConstraintNode struct {
	value   float64
	visited bool
}

type CPMap interface {
	Get(state State) (ConstraintNode, bool)
	Put(state State, value ConstraintNode)
	Clear()
}

type cheapestPathConstraint2 struct {
	m CPMap
}

func (c cheapestPathConstraint2) onExpand(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value < current.value {
		c.m.Put(node.state, ConstraintNode{node.value, false})
		return false
	}
	return true
}

func (c cheapestPathConstraint2) onVisit(node *node) bool {
	current, ok := c.m.Get(node.state)
	if !ok || node.value < current.value || node.value == current.value && !current.visited {
		c.m.Put(node.state, ConstraintNode{node.value, true})
		return false
	}
	return true
}

func (c cheapestPathConstraint2) reset() {
	c.m.Clear()
}

func CheapestPathConstraint2(m CPMap) Constraint {
	return cheapestPathConstraint2{m}
}
