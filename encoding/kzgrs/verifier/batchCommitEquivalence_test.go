package kzgrs_test

import (
	"testing"

	"github.com/Layr-Labs/eigenda/encoding/kzgrs"
	"github.com/Layr-Labs/eigenda/encoding/rs"
	"github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchEquivalence(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgrs.NewKzgEncoderGroup(kzgConfig, true)
	params := rs.GetEncodingParams(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)))
	enc, err := group.NewKzgEncoder(params)
	require.Nil(t, err)

	inputFr := rs.ToFrArray(GETTYSBURG_ADDRESS_BYTES)
	commit, g2commit, _, _, _, err := enc.Encode(inputFr)
	require.Nil(t, err)

	numBlob := 5
	commitPairs := make([]kzgrs.CommitmentPair, numBlob)
	for z := 0; z < numBlob; z++ {
		commitPairs[z] = kzgrs.CommitmentPair{
			Commitment:       *commit,
			LengthCommitment: *g2commit,
		}
	}

	assert.NoError(t, group.BatchVerifyCommitEquivalence(commitPairs), "batch equivalence negative test failed\n")

	var modifiedCommit bn254.G1Point
	bn254.AddG1(&modifiedCommit, commit, commit)
	for z := 0; z < numBlob; z++ {
		commitPairs[z] = kzgrs.CommitmentPair{
			Commitment:       modifiedCommit,
			LengthCommitment: *g2commit,
		}
	}

	assert.Error(t, group.BatchVerifyCommitEquivalence(commitPairs), "batch equivalence negative test failed\n")

	for z := 0; z < numBlob; z++ {
		commitPairs[z] = kzgrs.CommitmentPair{
			Commitment:       *commit,
			LengthCommitment: *g2commit,
		}
	}

	bn254.AddG1(&commitPairs[numBlob/2].Commitment, &commitPairs[numBlob/2].Commitment, &commitPairs[numBlob/2].Commitment)

	assert.Error(t, group.BatchVerifyCommitEquivalence(commitPairs), "batch equivalence negative test failed in outer loop\n")
}