# Solve

[![GoDoc](https://godoc.org/github.com/bertbaron/solve?status.svg)](https://godoc.org/github.com/bertbaron/solve)
[![Go Report Card](https://goreportcard.com/badge/github.com/bertbaron/solve)](https://goreportcard.com/report/github.com/bertbaron/solve)
[![Travis](https://travis-ci.org/bertbaron/solve.svg?branch=master)](https://travis-ci.org/bertbaron/solve)

## Go problem solving library with algorithms like A* (Astar), IDA* and Depth-first.

## Examples

There are several examples under the ```examples``` folder. Those can be run with ```go run```, for example:
```shell
$ go run examples/sudoku/main.go
```

Feel free to experiment with the examples.

## Usage

As an example we will sort a vector of elements with the minimal number of swaps
of neighboring elements:

```go
package main

import (
	"fmt"
	"github.com/bertbaron/solve"
)

type state struct {
	vector [5]byte
	cost   int
}

func (s state) Expand(ctx solve.Context) []solve.State {
	var children []solve.State
	for i := 0; i < len(s.vector)-1; i++ {
		copy := s.vector
		copy[i], copy[i+1] = copy[i+1], copy[i]
		children = append(children, state{copy, s.cost + 1})
	}
	return children
}

func (s state) IsGoal(ctx solve.Context) bool {
	for i := 1; i < len(s.vector); i++ {
		if s.vector[i-1] > s.vector[i] {
			return false
		}
	}
	return true
}

func (s state) Cost(ctx solve.Context) float64 {
	return float64(s.cost)
}

func (s state) Heuristic(ctx solve.Context) float64 {
	return 0
}

// Finds the minimum number of swaps of neighbouring elements required to
// sort a vector
func main() {
	s := state{[...]byte{3, 2, 5, 4, 1}, 0}
	result := solve.NewSolver(s).Solve()
	for _, st := range result.Solution {
		fmt.Printf("%v\n", st.(state).vector)
	}
	fmt.Printf("Visited %d nodes", result.Visited)
}
```
results in:
```
[3 2 5 4 1]
[3 2 5 1 4]
[3 2 1 5 4]
[2 3 1 5 4]
[2 1 3 5 4]
[1 2 3 5 4]
[1 2 3 4 5]
Visited 1406 nodes
```

The result contains statistics that tell for example how many nodes have been visited by
the algorithm. This can be quite useful during tuning.

The solution contains the complete path from initial state to goal state. If no solution
is found then result.Solution will yield an empty slice.

### Choosing the algorithm

#### A*
 
By default *A** is used to solve the problem. It can also explicitly be specified:

```go
        result := solve.NewSolver(s).
                Algorithm(solve.Astar).
                Solve()
```

A* will find the optimal solution in the least number of visited nodes if an admissible
heuristic is provided. The *heuristic function* estimates the remaining costs to reach
the goal for a given state. An admissible heuristic is a heuristic that never over-estimates
the actual costs of the best solution.

Since A* keeps all nodes in memory, it may run out of memory before a solution has been
found

#### IDA*

Iterative Deepening A*. Returns the optimal solution like A*, but uses
very little memory by using a depth first search in iterations with increasing limit. Works very well when the costs increase in discrete steps along the path.

```go
        result := solve.NewSolver(s).
                Algorithm(solve.IDAstar).
                Solve()
```
    
#### Depth First

Explores as far as possible along each branch before backtracking. Will not guarantee to
find the optimal solution. Requires very little memory (linear in the deepest path).

A limit can be provided to avoid that the search is expanding into depth forever.

```go
        result := solve.NewSolver(s).
                Algorithm(solve.DepthFirst).
                Limit(8).
                Solve()
```

#### Breadth First

Expands all nodes at a specific depth before going to the next
depth. Will find the optimal solution if the shortest path is the optimal solution.
Requires a lot of memory. It is almost always better to use A* or IDA*. If the cost
is proportional to the depth however and no heuristic is provided then breadth-first
is equivalent to A* but faster.

```go
        result := solve.NewSolver(s).
                Algorithm(solve.BreadthFirst).
                Solve()
```

### Tuning

We now have a program that can solve our problem and this may be all we need. However, if we
expect to solve the problem for larger input vectors we may run into trouble. Even other
vectors of length 5 will already show this:

```go
        s := state{[...]byte{5,4,3,2,1}, 0}
        result := solve.NewSolver(s).Solve()
        fmt.Printf("Visited %d nodes", result.Visited)
```
will result in:
```
Visited 350060 nodes
```
    
That seems like a huge number of nodes for such a small input. The reason is that the algorithm
has exponential complexity, and in case of A* also memory usage. Every state expands to 4 child-states.
The number of nodes at depth 10 is therefore 4^10=1048576.

It is almost always necessary to reduce the size of the search tree in order to be able to solve
slightly more complex problems. 

#### using heuristics

The most obvious and powerful way to reduce the search space when using A\* or IDA\* is of course the *heuristic function*.

For our problem an admissible heuristic could be the distance from each element to its target position,
divided by 2, since each step will only move two elements by one position each:

```go
func (s state) Heuristic(ctx solve.Context) float64 {
	goal := [5]byte{1,2,3,4,5} // would normally be precalculated
	n := len(goal)
	offset := 0
	for i, value := range s.vector {
		for d := 0; d < n; d++ {
			l, r := i - d, i + d
			if l >= 0 && goal[l] == value || r < n && goal[r] == value {
				offset += d
				break
			}
		}
	}
	return float64(offset / 2)
}
```
This results in
```
Visited 620 nodes
```

#### Using constraints

The easiest way to (further) reduce the size of the tree is by trying to see if one of the provided constraints
is suitable for the problem.

##### no-loop-constraint

Analyzing our problem domain, it is clear that it makes no sense to swap the same elements twice after
each other, since we will return to original state. We can use the no-loop-constraint to eliminate branches
if the state is equal to its parent state (not possible in this problem) or grandparents state.

For this we define a function that checks if two states are equal, and create a no-loop-constraint with a depth
of two

```go
func sameState(a, b solve.State) bool {
	return a.(state).vector == b.(state).vector
}
...
	result := solve.NewSolver(s).
		Constraint(solve.NoLoopConstraint(2, sameState)).
		Solve()
```
Even without heuritic this will result in:
```
Visited 39385 nodes
```

This saves an order of magnitude, since every expand will now effectively result in three child states
instead of four. This means we should now be able to compute vectors with one more element at the
same costs.

We can also detect bigger cycles by passing a greater depth:
```
	result := solve.NewSolver(s).
		Constraint(solve.NoLoopConstraint(100, sameState)).
		Solve()
```
```
Visited 28440 nodes
```

Note however that the costs will be linear in the depth (or depth of the search tree if that is
smaller), so the costs will increase while the benefit will typically decrease.

##### cheapest-path-constraint

The library also provides a constraint that detects if a state is reached via multiple paths, and
only continues with the best path to that state. This will for example detect that swapping elements
(1,2) and then (3,4) will result in the same state as swapping first (3,4) and then (1,2) and continue
with only one of the paths:

```go
        var stateMap cpMap
        result := solve.NewSolver(s).
                Constraint(solve.CheapestPathConstraint(&stateMap)).
                Solve()
        for _, st := range result.Solution {
                fmt.Printf("%v\n", st.(state).vector)
        }
```
    
This constraint basically turns the search tree into an actual graph of unique states. It works perfectly
for problems where the number of unique states is limited. The memory consumption is linear in
the number of unique states however.

A custom map implementation needs to be provided for efficient memory usage and performance. See
<https://godoc.org/github.com/bertbaron/solve#CPMap> for an example.

### Finding all solutions

After a solution have been found, a subsequent call to ```Solver.Solve()``` will continue the search.

Note that, in order to support this, the solver keeps the state of the search in memory until Solver.Completed() is
true. It is therefore reccommended to not keep a reference to the solver when it is not needed anymore so that it can
be garbage collected.


## License

Copyright © 2017

Distributed under the Eclipse Public License version 1.0
