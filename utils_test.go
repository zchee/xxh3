package xxh3

import (
	"math/rand"
	"testing"

	"github.com/zeebo/assert"
)

func TestInitSecretVariants(t *testing.T) {
	tests := map[string]struct {
		seed uint64
	}{
		"zero":     {seed: 0},
		"one":      {seed: 1},
		"high bit": {seed: 1 << 63},
		"maximum":  {seed: ^uint64(0)},
		"random":   {seed: rand.New(rand.NewSource(1)).Uint64()},
	}
	variants := map[string]func(ptr, u64){}
	if hasAVX2 {
		variants["AVX2"] = initSecretAVX2
	}
	if hasNEON {
		variants["NEON"] = func(secret ptr, seed u64) {
			initSecretNEON(secret, key, seed)
		}
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var want [secretSize]byte
			initSecret(ptr(&want), test.seed)
			for name, initialize := range variants {
				t.Run(name, func(t *testing.T) {
					// Sentinels catch writes outside the complete secret buffer.
					var guarded [secretSize + 2]byte
					for i := range guarded {
						guarded[i] = 0xa5
					}
					initialize(ptr(&guarded[1]), test.seed)
					assert.Equal(t, guarded[1:secretSize+1], want[:])
					assert.Equal(t, guarded[0], byte(0xa5))
					assert.Equal(t, guarded[secretSize+1], byte(0xa5))
				})
			}
		})
	}
}
