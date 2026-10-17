package xxh3

import (
	"encoding/binary"
	"os"
	"testing"
)

const (
	goldenMaxLength = 2400
	goldenSeed      = uint64(0x9e3779b97f4a7c15)
	goldenRowSize   = 6 * 8
)

// goldenData uses an LCG starting at 1, with multiplier 6364136223846793005
// and increment 1442695040888963407, taking the high byte after each update.
func goldenData() []byte {
	data := make([]byte, goldenMaxLength)
	state := uint64(1)
	for i := range data {
		state = state*6364136223846793005 + 1442695040888963407
		data[i] = byte(state >> 56)
	}
	return data
}

func TestGoldenLengths(t *testing.T) {
	fixture, err := os.ReadFile("testdata/golden_lengths.bin")
	if err != nil {
		t.Fatal(err)
	}
	if want := (goldenMaxLength + 1) * goldenRowSize; len(fixture) != want {
		t.Fatalf("golden fixture size: got %d, want %d", len(fixture), want)
	}
	data := goldenData()
	tests := map[string]struct {
		available bool
		with      func(func())
	}{
		"auto":    {true, func(check func()) { check() }},
		"generic": {true, withGeneric},
		"avx512":  {hasAVX512, withAVX512},
		"avx2":    {hasAVX2, withAVX2},
		"sse2":    {hasSSE2, withSSE2},
	}
	for name, tt := range tests {
		if !tt.available {
			continue
		}
		t.Run(name, func(t *testing.T) {
			tt.with(func() {
				for n := range goldenMaxLength + 1 {
					input := data[:n]
					h128 := Hash128(input)
					h128seed := Hash128Seed(input, goldenSeed)
					got := [6]uint64{Hash(input), HashSeed(input, goldenSeed), h128.Lo, h128.Hi, h128seed.Lo, h128seed.Hi}
					var want [6]uint64
					for i := range want {
						want[i] = binary.LittleEndian.Uint64(fixture[n*goldenRowSize+i*8:])
					}
					if got != want {
						t.Fatalf("length %d: [Hash, HashSeed(%#x), Hash128.Lo, Hash128.Hi, Hash128Seed.Lo, Hash128Seed.Hi] = %#x, want %#x", n, goldenSeed, got, want)
					}
				}
			})
		})
	}
}
