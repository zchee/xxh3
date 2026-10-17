//go:build arm64

package xxh3

import "testing"

func TestAccumNEONInitParity(t *testing.T) {
	data := goldenData(2400 + 15)
	initial := [8]u64{
		prime32_3, prime64_1, prime64_2, prime64_3,
		prime64_4, prime32_2, prime64_5, prime32_1,
	}
	tests := map[string]struct {
		seed uint64
	}{
		"default":   {seed: 0},
		"seed one":  {seed: 1},
		"seed high": {seed: 1 << 63},
		"seed max":  {seed: ^uint64(0)},
		"seed mix":  {seed: 0x93b169c748af250e},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var secret [secretSize]byte
			k := key
			if tt.seed != 0 {
				initSecret(ptr(&secret), tt.seed)
				k = ptr(&secret)
			}
			for n := 241; n <= 2400; n++ {
				for _, offset := range []int{0, 1, 8, 15} {
					want := initial
					got := [8]u64{1, 3, 5, 7, 11, 13, 17, 19}
					p := ptr(&data[offset])
					accumNEON(&want, p, k, u64(n))
					accumNEONInit(&got, p, k, u64(n))
					if got != want {
						t.Fatalf("length=%d offset=%d: got %016x, want %016x", n, offset, got, want)
					}
				}
			}
		})
	}
	t.Run("zero length", func(t *testing.T) {
		got := [8]u64{1, 3, 5, 7, 11, 13, 17, 19}
		accumNEONInit(&got, nil, nil, 0)
		if got != initial {
			t.Fatalf("got %016x, want initial accumulators %016x", got, initial)
		}
	})
}
