package xxh3

import (
	"encoding/binary"
	"os"
	"testing"
)

// TestGenerateGoldenLengths must only run in an isolated d0c0a56 reference tree.
// Set XXH3_GOLDEN_GEN=1 and select this test explicitly to generate the fixture.
func TestGenerateGoldenLengths(t *testing.T) {
	if os.Getenv("XXH3_GOLDEN_GEN") != "1" {
		t.Skip("set XXH3_GOLDEN_GEN=1 in the isolated reference tree to generate the fixture")
	}
	data := goldenData(goldenMaxLength)
	fixture := make([]byte, 0, (goldenMaxLength+1)*goldenRowSize)
	for n := range goldenMaxLength + 1 {
		input := data[:n]
		h128 := Hash128(input)
		h128seed := Hash128Seed(input, goldenSeed)
		// Each row is six little-endian uint64s: 64-bit default and seeded,
		// then the low and high words of 128-bit default and seeded hashes.
		for _, value := range [...]uint64{Hash(input), HashSeed(input, goldenSeed), h128.Lo, h128.Hi, h128seed.Lo, h128seed.Hi} {
			fixture = binary.LittleEndian.AppendUint64(fixture, value)
		}
	}
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/golden_lengths.bin", fixture, 0o644); err != nil {
		t.Fatal(err)
	}
}
