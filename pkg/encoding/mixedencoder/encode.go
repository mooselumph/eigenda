package mixedencoder

import (
	"fmt"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	kzgrs "github.com/Layr-Labs/eigenda/pkg/encoding/kzgEncoder"
	bls "github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
)

type MixedEncoder struct {

	// KZG Encoder
	KzgEncoderGroup *kzgrs.KzgEncoderGroup
}

func NewMixedEncoder(kzgEncoderGroup *kzgrs.KzgEncoderGroup) *MixedEncoder {
	return &MixedEncoder{
		KzgEncoderGroup: kzgEncoderGroup,
	}
}

type MixedEncodingOutput struct {
	Param         rs.EncodingParams
	Allocation    Allocation
	ShiftedCommit *bls.G1Point
	Frames        []kzgrs.Frame
	Indices       []uint32
}

func (e *MixedEncoder) Encode(input []byte, params []rs.EncodingParams) (*bls.G1Point, *bls.G1Point, []*MixedEncodingOutput, error) {

	coeffs := rs.ToFrArray(input)

	// Get Offsets
	allocations := make([]*Allocation, len(params))
	for i, param := range params {
		allocations[i] = &Allocation{
			NumEvaluations: int(param.NumChunks * param.ChunkLen),
		}
	}

	err := AddOffsets(allocations)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not add offsets: %w", err)
	}

	outputs := make([]*MixedEncodingOutput, len(params))

	var commit, lengthProof *bls.G1Point

	// Encode
	for ind, param := range params {

		// Get Encoder
		encoder, err := e.KzgEncoderGroup.NewKzgEncoder(param)
		if err != nil {
			return nil, nil, nil, err
		}

		// Compute Commitment
		if commit == nil {
			commit = encoder.Commit(coeffs)
			lengthProof = encoder.GenerateLengthProof(coeffs)
		}

		// Condition the input
		shiftedPolyCoeffs := ShiftPoly(coeffs, allocations[ind].Offset)

		// Encode
		shiftedCommit, _, frames, indices, err := encoder.Encode(shiftedPolyCoeffs)
		if err != nil {
			return nil, nil, nil, err
		}

		outputs[ind] = &MixedEncodingOutput{
			Param:         param,
			Frames:        frames,
			Indices:       indices,
			ShiftedCommit: shiftedCommit,
			Allocation:    *allocations[ind],
		}

	}

	return commit, lengthProof, outputs, nil
}

func ShiftPoly(coeffs []bls.Fr, factor bls.Fr) []bls.Fr {

	shift := &bls.Fr{}
	bls.CopyFr(shift, &factor)

	shiftedCoeffs := make([]bls.Fr, len(coeffs))

	bls.CopyFr(&shiftedCoeffs[0], &coeffs[0])
	for i := 1; i < len(coeffs); i++ {
		bls.MulModFr(&shiftedCoeffs[i], &coeffs[i], shift)
		bls.MulModFr(shift, shift, &factor)
	}

	return shiftedCoeffs
}
