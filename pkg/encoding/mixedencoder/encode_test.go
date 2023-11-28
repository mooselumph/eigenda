package mixedencoder_test

import (
	"fmt"
	"runtime"
	"testing"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	kzgrs "github.com/Layr-Labs/eigenda/pkg/encoding/kzgEncoder"
	"github.com/Layr-Labs/eigenda/pkg/encoding/mixedencoder"
	"github.com/Layr-Labs/eigenda/pkg/kzg/bn254"
	"github.com/stretchr/testify/assert"
)

var (
	group *kzgrs.KzgEncoderGroup

	gettysburgAddressBytes = []byte("Fourscore and seven years ago our fathers brought forth, on this continent, a new nation, conceived in liberty, and dedicated to the proposition that all men are created equal. Now we are engaged in a great civil war, testing whether that nation, or any nation so conceived, and so dedicated, can long endure. We are met on a great battle-field of that war. We have come to dedicate a portion of that field, as a final resting-place for those who here gave their lives, that that nation might live. It is altogether fitting and proper that we should do this. But, in a larger sense, we cannot dedicate, we cannot consecrate—we cannot hallow—this ground. The brave men, living and dead, who struggled here, have consecrated it far above our poor power to add or detract. The world will little note, nor long remember what we say here, but it can never forget what they did here. It is for us the living, rather, to be dedicated here to the unfinished work which they who fought here have thus far so nobly advanced. It is rather for us to be here dedicated to the great task remaining before us—that from these honored dead we take increased devotion to that cause for which they here gave the last full measure of devotion—that we here highly resolve that these dead shall not have died in vain—that this nation, under God, shall have a new birth of freedom, and that government of the people, by the people, for the people, shall not perish from the earth.")
)

func setup(t *testing.T) {

	kzgConfig := &kzgrs.KzgConfig{
		G1Path:    "../../../inabox/resources/kzg/g1.point.300000",
		G2Path:    "../../../inabox/resources/kzg/g2.point.300000",
		CacheDir:  "../../../inabox/resources/kzg/SRSTables",
		SRSOrder:  300000,
		NumWorker: uint64(runtime.GOMAXPROCS(0)),
	}

	var err error
	group, err = kzgrs.NewKzgEncoderGroup(kzgConfig)
	if err != nil {
		t.Fatal(err)
	}

}

func TestMixedEncoding(t *testing.T) {

	setup(t)

	// Make the mixed encoder
	encoder := mixedencoder.NewMixedEncoder(group)

	// Make the input

	// blobLength := 256
	// blob := make([]byte, blobLength*31)
	// _, err := rand.Read(blob)
	// assert.NoError(t, err)
	blob := gettysburgAddressBytes

	// Make the params
	params := []rs.EncodingParams{
		{
			NumChunks: 128,
			ChunkLen:  8,
		},
		{
			NumChunks: 32,
			ChunkLen:  64,
		},
		{
			NumChunks: 1,
			ChunkLen:  1024,
		},
	}

	// Encode
	commit, _, outputs, err := encoder.Encode(blob, params)
	assert.NoError(t, err)

	_ = commit
	// Check the proofs
	// for _, output := range outputs {
	// 	verifyFrames(t, commit, output)
	// }

	// Decode
	inputs := make([]*mixedencoder.MixedDecoderInput, len(outputs))
	for i, output := range outputs {

		frames := make([]rs.Frame, len(output.Frames))
		for j, frame := range output.Frames {
			frames[j] = rs.Frame{
				Coeffs: frame.Coeffs,
			}
		}

		inputs[i] = &mixedencoder.MixedDecoderInput{
			EncodingParams: output.Param,
			Allocation:     output.Allocation,
			Frames:         frames,
			Indices:        output.Indices,
		}
	}

	numEvaluations := 0
	for _, input := range inputs {
		numEvaluations += input.Allocation.NumEvaluations
	}
	numEvaluations = int(rs.NextPowerOf2(uint64(numEvaluations)))

	tInputs := []*mixedencoder.MixedDecoderInput{
		inputs[1], inputs[2],
	}

	decoded, err := encoder.Decode(numEvaluations, len(blob), tInputs)

	assert.NoError(t, err)
	assert.Equal(t, string(blob), string(decoded))

	fmt.Println("Offset", tInputs[0].Allocation.RootIndex)
	fmt.Println(string(decoded))

}

func verifyFrames(t *testing.T, commit *bn254.G1Point, output *mixedencoder.MixedEncodingOutput) {

	frames := output.Frames
	indices := output.Indices

	enc, err := group.GetKzgEncoder(output.Param)
	if err != nil {
		t.Fatalf("Error making rs: %q", err)
	}

	for _, frame := range frames {
		assert.NotEqual(t, len(frame.Coeffs), 0)
	}

	for i := 0; i < len(frames); i++ {
		f := frames[i]
		j := indices[i]

		q, err := rs.GetLeadingCosetIndex(uint64(i), output.Param.NumChunks)
		assert.Nil(t, err)

		assert.Equal(t, j, q, "leading coset inconsistency")

		fmt.Printf("frame %v leading coset %v\n", i, j)
		lc := enc.Fs.ExpandedRootsOfUnity[uint64(q)]

		assert.True(t, f.Verify(enc.Ks, output.ShiftedCommit, &lc), "Proof %v failed\n", i)
	}

}
