package kd

import (
	"github.com/BenLubar/goray/geometry"
	"sort"
)

const (
	DIM_X = 0
	DIM_Y = 1
	DIM_Z = 2
)

// The KDNodes are the nodes in the tree
// It has a value, a splitting dimension and left and right childs.
type KDNode struct {
	Position    geometry.Vec3
	Split       int
	Left, Right *KDNode
}

// Convinience distance function
func (me *KDNode) Distance(other geometry.Vec3) float64 {
	return me.Position.Distance(other)
}

// Convinience distance^2 function
func (me *KDNode) Distance2(other geometry.Vec3) float64 {
	return me.Position.Distance2(other)
}

// Extract the correct value from the geometry.Vec3 to compare on
func comparingValue(item geometry.Vec3, dimension int) float64 {
	switch dimension {
	case DIM_X:
		return item.X
	case DIM_Y:
		return item.Y
	case DIM_Z:
		return item.Z
	}
	panic("Trying to get higher dimensional value")
}

type valueList struct {
	values    []geometry.Vec3
	dimension int
}

func (l valueList) Len() int {
	return len(l.values)
}

func (l valueList) Less(i, j int) bool {
	return comparingValue(l.values[i], l.dimension) < comparingValue(l.values[j], l.dimension)
}

func (l valueList) Swap(i, j int) {
	l.values[i], l.values[j] = l.values[j], l.values[i]
}

// Debugging functions calculating the KD tree depth
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (k *KDNode) Depth() int {
	if k == nil {
		return 0
	}
	depth := max(k.Left.Depth(), k.Right.Depth())

	return 1 + depth
}

// Creates a new KD-tree by taking a *list.List of KDValues
// Works by finding the median in every dimension and
// recursivly creating KD-trees as children untill the list is empty.
//
// Uses Go routines and channels to acheive concurrency.
// Every level creates one new Go routine and processes one sub-tree
// on it's own.
func New(items []geometry.Vec3) *KDNode {
	return create(items, 0)
}

func create(l []geometry.Vec3, depth int) *KDNode {
	if len(l) == 0 {
		return nil
	}

	// Sort the array
	sort.Sort(valueList{l, depth % 3})
	median := len(l) / 2
	// Adjust the median to make sure it's the FIRST of any
	// identical values
	dimension := depth % 3
	forbiddenValue := comparingValue(l[median], dimension)
	for comparingValue(l[median], dimension) == forbiddenValue && median > 0 {
		median--
	}
	value := l[median]

	left := create(l[:median], depth+1)
	right := create(l[median+1:], depth+1)

	return &KDNode{value, dimension, left, right}
}

// Searches the tree for any nodes within radius r
// from the target point. This is currently rather slow
// but accurate. By comparing every point to the leftmost
// and rightmost point to the resulting sphere
// irrelevant subtrees are cut of.
func (tree *KDNode) Neighbors(point geometry.Vec3, r float64) []*KDNode {
	return tree.neighbors(point, r, nil)
}

func (tree *KDNode) neighbors(point geometry.Vec3, r float64, result []*KDNode) []*KDNode {
	if tree == nil {
		return result
	}

	// Am I part of the sphere?
	// Compare Distance² to r² to avoid calling sqrt
	if tree.Distance2(point) < r*r {
		result = append(result, tree)
	}

	split := tree.Split
	// Is the leftmost point to the left of us?
	if comparingValue(tree.Position, split) > comparingValue(point, split)-r {
		result = tree.Left.neighbors(point, r, result)
	}

	// Is the rightmost point to the right of us?
	if comparingValue(tree.Position, split) < comparingValue(point, split)+r {
		result = tree.Right.neighbors(point, r, result)
	}

	// Return all the found nodes
	return result
}
