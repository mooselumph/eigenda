package kzgEncoder_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	rs "github.com/Layr-Labs/eigenda/pkg/encoding/encoder"
	kzgRs "github.com/Layr-Labs/eigenda/pkg/encoding/kzgEncoder"
	"github.com/stretchr/testify/assert"
)

func TestBenchmarkEncoding(t *testing.T) {
	// t.Skip("This test is meant to be run manually, not as part of the test suite")
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)

	// chunkLengths := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
	// chunkCounts := []int{4, 8, 16}

	chunkLengths := []int{64, 128, 256, 512, 1024, 2048, 4096, 8192}
	chunkCounts := []int{4, 8, 16}

	file, err := os.Create("benchmark_results.csv")
	if err != nil {
		t.Fatalf("Failed to open file for writing: %v", err)
	}
	defer file.Close()

	// fmt.Fprintln(file, "numChunks,chunkLength,ns/op,allocs/op")

	for _, chunkLength := range chunkLengths {

		// blobSize := chunkLength * 31 * 2
		// blob := make([]byte, blobSize)
		blob := GETTYSBURG_ADDRESS_BYTES
		_, err = rand.Read(blob)
		assert.NoError(t, err)

		fmt.Println("BLoblength: ", len(blob))

		for _, numChunks := range chunkCounts {

			params := rs.EncodingParams{
				ChunkLen:  uint64(chunkLength),
				NumChunks: uint64(numChunks),
			}

			fmt.Printf("NumChunks: %d, ChunkLength: %d \n", numChunks, chunkLength)
			benchmarkEncoding(t, group, blob, params, false)

			result := testing.Benchmark(func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					// control = profile.Start(profile.ProfilePath("."))

					fmt.Printf("Benchmarking encoding with %d chunks of length %d", numChunks, chunkLength)
					benchmarkEncoding(t, group, blob, params, false)

					// control.Stop()
				}
			})
			// Print results in CSV format
			fmt.Fprintf(file, "%d,%d,%d,%d\n", numChunks, chunkLength, result.NsPerOp(), result.AllocsPerOp())

		}
	}

}

func benchmarkEncoding(t *testing.T, group *kzgRs.KzgEncoderGroup, input []byte, params rs.EncodingParams, decode bool) {

	enc, err := group.GetKzgEncoder(params)
	if err != nil {
		t.Errorf("Error making rs: %q", err)
	}

	//encode the data
	_, _, frames, _, err := enc.EncodeBytes(input)

	if err != nil {
		t.Errorf("Error Encoding:\n Data:\n %q \n Err: %q", input, err)
	}

	for _, frame := range frames {
		assert.NotEqual(t, len(frame.Coeffs), 0)
	}

	if !decode {
		return
	}

	//sample the correct systematic frames
	samples, indices := sampleFrames(frames, uint64(len(frames)))

	data, err := enc.Decode(samples, indices, uint64(len(input)))
	if err != nil {
		t.Errorf("Error Decoding:\n Data:\n %q \n Err: %q", input, err)
	}
	assert.Equal(t, input, data, "Input data was not equal to the decoded data")

}

func TestParallelEncoding(t *testing.T) {

	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)

	params := []rs.EncodingParams{
		{
			ChunkLen:  512,
			NumChunks: 128,
		},
		{
			ChunkLen:  64,
			NumChunks: 1024,
		},
		{
			ChunkLen:  8,
			NumChunks: 8192,
		},
	}

	blobLength := 2000
	blob := make([]byte, blobLength*31)
	_, err := rand.Read(blob)
	assert.NoError(t, err)

	numRoutines := 1

	for _, param := range params {

		start := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				benchmarkEncoding(t, group, blob, param, false)
			}()
		}

		wg.Wait()

		fmt.Println("Parallel Encoding took ", time.Since(start))

	}

}
