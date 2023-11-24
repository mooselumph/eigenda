package kzgEncoder_test

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	kzgRs "github.com/Layr-Labs/eigenda/pkg/encoding/kzgEncoder"
	bls "github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
	"github.com/stretchr/testify/assert"
)

var (
	codingRate = 5
	q          = 8

	d = distribution{
		smallStake:  32,
		mediumStake: 5000,
		largeStake:  30000,

		numSmall:  100,
		numMedium: 30,
		numLarge:  8,
	}
)

type distribution struct {
	smallStake  int
	mediumStake int
	largeStake  int

	numSmall  int
	numMedium int
	numLarge  int
}

func (d distribution) totalStake() int {
	return d.smallStake + d.mediumStake + d.largeStake
}

func (d distribution) numTotal() int {
	return d.numSmall + d.numMedium + d.numLarge
}

func getChunkSize(encodedBlobSize, stake, totalStake int) int {
	return int(rs.NextPowerOf2(uint64(encodedBlobSize * stake / totalStake)))
}

func getNumChunks(encodedBlobSize, chunkSize, stake, totalStake int) int {

	return encodedBlobSize*stake/(totalStake*chunkSize) + 1

}

func getMixedEncodingParams(d distribution, encodedBlobSize int) []rs.EncodingParams {
	return []rs.EncodingParams{
		{
			NumChunks: rs.NextPowerOf2(uint64(d.numSmall)),
			ChunkLen:  uint64(getChunkSize(encodedBlobSize, d.smallStake, d.totalStake())),
		},
		{
			NumChunks: rs.NextPowerOf2(uint64(d.numMedium)),
			ChunkLen:  uint64(getChunkSize(encodedBlobSize, d.mediumStake, d.totalStake())),
		},
		{
			NumChunks: rs.NextPowerOf2(uint64(d.numLarge)),
			ChunkLen:  uint64(getChunkSize(encodedBlobSize, d.largeStake, d.totalStake())),
		},
	}
}

func getStandardEncodingParams(d distribution, encodedBlobSize int) []rs.EncodingParams {

	chunkSize := int(rs.NextPowerOf2(uint64(encodedBlobSize/(d.numTotal()*q) + 1)))

	numChunks := d.numSmall * getNumChunks(encodedBlobSize, chunkSize, d.smallStake, d.totalStake())
	numChunks += d.numMedium * getNumChunks(encodedBlobSize, chunkSize, d.mediumStake, d.totalStake())
	numChunks += d.numLarge * getNumChunks(encodedBlobSize, chunkSize, d.largeStake, d.totalStake())

	numChunks = int(rs.NextPowerOf2(uint64(numChunks)))

	return []rs.EncodingParams{
		{
			NumChunks: uint64(numChunks),
			ChunkLen:  uint64(chunkSize),
		},
	}

}

func TestMixedChunkSizes(t *testing.T) {

	// t.Skip("This test is meant to be run manually, not as part of the test suite")
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)

	blobSizes := []int{1000}

	for _, blobSize := range blobSizes {

		blob := make([]byte, blobSize)
		_, err := rand.Read(blob)
		assert.NoError(t, err)

		// Test Mixed Encoding
		params := getMixedEncodingParams(d, blobSize*codingRate)
		testMultiEncoding(t, group, params, blob)

		start := time.Now()
		testMultiEncoding(t, group, params, blob)
		elapsed := time.Since(start)
		t.Logf("Mixed Encoding took %s", elapsed)

		// Test Standard Encoding
		params = getStandardEncodingParams(d, blobSize*codingRate)
		testMultiEncoding(t, group, params, blob)

		start = time.Now()
		testMultiEncoding(t, group, params, blob)
		elapsed = time.Since(start)
		t.Logf("Standard Encoding took %s", elapsed)

	}

}

var (
	output []bls.G1Point
)

func testMultiEncoding(t *testing.T, group *kzgRs.KzgEncoderGroup, params []rs.EncodingParams, blob []byte) {

	for _, param := range params {

		fmt.Printf("NumChunks: %d, ChunkLength: %d \n", param.NumChunks, param.ChunkLen)

		enc, err := group.NewKzgEncoder(param)
		if err != nil {
			t.Fatalf("Error making rs (chunkLength %v, numChunks %v): %q", param.ChunkLen, param.NumChunks, err)
		}

		//generate the proofs

		coeffs := rs.ToFrArray(blob)
		paddedCoeffs := make([]bls.Fr, enc.NumEvaluations())
		copy(paddedCoeffs, coeffs)

		proofs, err := enc.ProveAllCosetThreads(paddedCoeffs, enc.NumChunks, enc.ChunkLen, enc.NumWorker)
		if err != nil {
			t.Errorf("Error making rs: %q", err)
		}

		output = proofs

	}
}
