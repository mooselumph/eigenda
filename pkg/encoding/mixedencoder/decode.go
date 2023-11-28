package mixedencoder

import (
	"fmt"
	"math"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	"github.com/Layr-Labs/eigenda/pkg/kzg"
	bls "github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
)

type MixedDecoderInput struct {
	rs.EncodingParams
	Allocation Allocation
	Frames     []rs.Frame
	Indices    []uint32
}

func (e *MixedEncoder) Decode(numEvaluations, inputSize int, inputs []*MixedDecoderInput) ([]byte, error) {

	samples := make([]*bls.Fr, numEvaluations)

	for _, input := range inputs {

		interval := numEvaluations / input.Allocation.NumEvaluations

		// Get Encoder
		enc, err := rs.NewEncoder(input.EncodingParams, false)
		if err != nil {
			return nil, err
		}

		// copy evals based on frame coeffs into samples
		for i, d := range input.Indices {
			f := input.Frames[i]
			e, err := rs.GetLeadingCosetIndex(uint64(d), input.NumChunks)
			if err != nil {
				return nil, err
			}

			evals, err := enc.GetInterpolationPolyEval(f.Coeffs, uint32(e))
			if err != nil {
				return nil, err
			}

			// Some pattern i butterfly swap. Find the leading coset, then increment by number of coset
			for j := uint64(0); j < input.ChunkLen; j++ {
				p := j*input.NumChunks + uint64(e)

				// Apply the interval and offset
				p = p*uint64(interval) + uint64(input.Allocation.RootIndex)

				if samples[p] != nil {
					return nil, fmt.Errorf("duplicate index %v", p)
				}
				samples[p] = new(bls.Fr)
				bls.CopyFr(samples[p], &evals[j])
			}
		}

	}

	reconstructedData := make([]bls.Fr, numEvaluations)
	missingIndices := false
	for i, s := range samples {
		if s == nil {
			missingIndices = true
			break
		}
		reconstructedData[i] = *s
	}

	n := uint8(math.Log2(float64(numEvaluations)))
	fs := kzg.NewFFTSettings(n)

	if missingIndices {
		var err error
		reconstructedData, err = fs.RecoverPolyFromSamples(
			samples,
			fs.ZeroPolyViaMultiplication,
		)
		if err != nil {
			return nil, err
		}
	}

	reconstructedPoly, err := fs.FFT(reconstructedData, true)
	if err != nil {
		return nil, err
	}

	data := rs.ToByteArray(reconstructedPoly, uint64(inputSize))

	return data, nil
}
