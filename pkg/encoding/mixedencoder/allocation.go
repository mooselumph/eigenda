package mixedencoder

import (
	"fmt"
	"math"
	"sort"

	kzg "github.com/Layr-Labs/eigenda/pkg/kzg"
	bls "github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
)

type TreeNode struct {
	Value  int
	Offset int
	Left   *TreeNode
	Right  *TreeNode
}

type BinaryTree struct {
	Root *TreeNode
}

func NewTree(depth uint8) *BinaryTree {

	root := NewNode(int(1<<depth), 0)
	buildTree(0, depth, root)
	return &BinaryTree{Root: root}
}

func buildTree(depth, maxDepth uint8, root *TreeNode) {
	if depth == maxDepth {
		return
	}

	depth++
	value := int(1 << (maxDepth - depth))
	root.Left = NewNode(value, root.Offset)
	root.Right = NewNode(value, root.Offset+1<<(depth-1))

	buildTree(depth, maxDepth, root.Left)
	buildTree(depth, maxDepth, root.Right)
}

func NewNode(value int, offset int) *TreeNode {
	return &TreeNode{Value: value, Offset: offset}
}

type Allocation struct {
	NumEvaluations int
	RootIndex      int
	Offset         bls.Fr
}

func AddOffsets(allocations []*Allocation) error {

	// Sort allocations by number of evaluations
	sorted := make([]*Allocation, len(allocations))
	copy(sorted, allocations)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].NumEvaluations > sorted[j].NumEvaluations
	})

	// Get the total number of evaluations
	totalEvaluations := 0
	for _, alloc := range sorted {
		totalEvaluations += alloc.NumEvaluations
	}
	depth := uint8(math.Ceil(math.Log2(float64(totalEvaluations))))

	fs := kzg.NewFFTSettings(depth)
	rootsOfUnity := fs.ExpandedRootsOfUnity

	// Create a tree with the total number of evaluations
	tree := NewTree(depth)

	remaining := dfsAssign(tree.Root, sorted, rootsOfUnity)
	if len(remaining) != 0 {
		return fmt.Errorf("could not assign all allocations")
	}

	return nil

}

func dfsAssign(node *TreeNode, allocations []*Allocation, rootsOfUnity []bls.Fr) []*Allocation {
	if node == nil {
		return allocations
	}

	if len(allocations) == 0 {
		return nil
	}

	// Assign the offset to the allocations
	if node.Value == allocations[0].NumEvaluations {
		allocations[0].RootIndex = node.Offset
		allocations[0].Offset = rootsOfUnity[node.Offset]
		allocations = allocations[1:]

		return allocations
	}

	allocations = dfsAssign(node.Left, allocations, rootsOfUnity)
	allocations = dfsAssign(node.Right, allocations, rootsOfUnity)
	return allocations
}
