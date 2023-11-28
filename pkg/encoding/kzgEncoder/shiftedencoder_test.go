package kzgEncoder_test

import (
	"crypto/rand"
	"fmt"
	"math"
	"testing"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	kzgRs "github.com/Layr-Labs/eigenda/pkg/encoding/kzgEncoder"
	kzg "github.com/Layr-Labs/eigenda/pkg/kzg"
	bls "github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
	"github.com/stretchr/testify/assert"
)

func TestShiftedEncoding(t *testing.T) {

	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)

	blobSize := 32
	blob := make([]byte, blobSize*31)
	_, err := rand.Read(blob)
	assert.NoError(t, err)

	input := rs.ToFrArray(blob)

	params := rs.EncodingParams{
		NumChunks: 4,
		ChunkLen:  16,
	}
	enc, err := group.NewKzgEncoder(params)
	if err != nil {
		t.Errorf("Error making rs: %q", err)
	}

	n := uint8(math.Log2(float64(enc.NumEvaluations()))) + 1
	fs := kzg.NewFFTSettings(n)

	factor := &bls.Fr{}
	shift := &bls.Fr{}
	bls.CopyFr(shift, fs.RootOfUnity)
	bls.CopyFr(factor, fs.RootOfUnity)
	for i := 1; i < len(input); i++ {
		bls.MulModFr(&input[i], &input[i], shift)
		bls.MulModFr(shift, shift, factor)
	}

	//encode the data
	_, _, frames, indices_, err := enc.Encode(input)
	if err != nil {
		t.Errorf("Error Encoding:\n Data:\n %q \n Err: %q", input, err)
	}

	fmt.Println("indices_", indices_)

	// for _, frame := range frames {
	// 	assert.NotEqual(t, len(frame.Coeffs), 0)
	// }

	// for i := 0; i < len(frames); i++ {
	// 	f := frames[i]
	// 	j := indices[i]

	// 	q, err := rs.GetLeadingCosetIndex(uint64(i), params.NumChunks)
	// 	assert.Nil(t, err)

	// 	assert.Equal(t, j, q, "leading coset inconsistency")

	// 	fmt.Printf("frame %v leading coset %v\n", i, j)
	// 	lc := enc.Fs.ExpandedRootsOfUnity[uint64(q)]

	// 	assert.True(t, f.Verify(enc.Ks, &shiftedCommit, &lc), "Proof %v failed\n", i)
	// }

	samples_, indices := sampleFrames(frames, uint64(len(frames)))

	samples := make([]rs.Frame, len(frames))
	for i, frame := range samples_ {
		samples[i] = rs.Frame{
			Coeffs: frame.Coeffs,
		}
	}

	fmt.Println("len(samples)", len(samples), "len(frames)", len(frames), "len(indices)", len(indices))

	recoveredCoeffs, err := enc.Encoder.Decode(samples, indices)
	assert.NoError(t, err)

	notEqual := make([]int, 0)
	for i := 0; i < len(input); i++ {
		if !bls.EqualFr(&input[i], &recoveredCoeffs[i]) {
			notEqual = append(notEqual, i)
		}
	}
	assert.Equal(t, []int{}, notEqual)

}
