package mixedencoder_test

import (
	"fmt"
	"testing"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	mixed "github.com/Layr-Labs/eigenda/pkg/encoding/mixedencoder"
	"github.com/stretchr/testify/assert"
)

func TestBuildTree(t *testing.T) {

	// Build the tree
	tree := mixed.NewTree(3)

	// Check the tree
	assert.Equal(t, 8, tree.Root.Value)
	assert.Equal(t, 0, tree.Root.Offset)

	assert.Equal(t, 4, tree.Root.Left.Value)
	assert.Equal(t, 0, tree.Root.Left.Offset)
	assert.Equal(t, 4, tree.Root.Right.Value)
	assert.Equal(t, 1, tree.Root.Right.Offset)

	assert.Equal(t, 2, tree.Root.Left.Left.Value)
	assert.Equal(t, 0, tree.Root.Left.Left.Offset)
	assert.Equal(t, 2, tree.Root.Left.Right.Value)
	assert.Equal(t, 2, tree.Root.Left.Right.Offset)
	assert.Equal(t, 2, tree.Root.Right.Left.Value)
	assert.Equal(t, 1, tree.Root.Right.Left.Offset)
	assert.Equal(t, 2, tree.Root.Right.Right.Value)
	assert.Equal(t, 3, tree.Root.Right.Right.Offset)

}

type assignment struct {
	numEvals  int
	rootIndex int
	index     int
}

func TestAddOffsets(t *testing.T) {

	// Build Allocations
	// numEvaluations := []int{1, 2, 2, 4, 4, 8, 8}
	// numEvaluations := []int{4, 8, 8, 16, 32}
	numEvaluations := []int{4, 8, 8, 16, 32, 64, 128, 256, 512, 1024}
	allocations := make([]*mixed.Allocation, len(numEvaluations))

	for ind, num := range numEvaluations {
		allocations[ind] = &mixed.Allocation{
			NumEvaluations: num,
		}
	}

	// Add offsets
	err := mixed.AddOffsets(allocations)
	assert.NoError(t, err)

	// Check ordering
	for ind, alloc := range allocations {
		assert.Equal(t, numEvaluations[ind], alloc.NumEvaluations)
	}

	// Check uniqueness of assignments
	totalEvaluations := 0
	for _, num := range numEvaluations {
		totalEvaluations += num
	}
	totalEvaluations = int(rs.NextPowerOf2(uint64(totalEvaluations)))
	fmt.Println("totalEvaluations", totalEvaluations)

	assignments := make([]*assignment, totalEvaluations)
	for _, alloc := range allocations {

		for i := 0; i < alloc.NumEvaluations; i++ {

			interval := totalEvaluations / alloc.NumEvaluations

			ind := alloc.RootIndex + i*interval

			if assignments[ind] != nil {
				t.Errorf("index %d assigned twice (root index %v, index %v, evals %v) and (root index %v, index %v, evals %v)",
					ind, alloc.RootIndex, i, alloc.NumEvaluations,
					assignments[ind].rootIndex, assignments[ind].index, assignments[ind].numEvals,
				)
			}
			assignments[ind] = &assignment{
				numEvals:  alloc.NumEvaluations,
				rootIndex: alloc.RootIndex,
				index:     i,
			}
		}
	}

}
