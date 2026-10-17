package xxh3

import "testing"

func TestSumTailLengths(t *testing.T) {
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
				for mode, tc := range map[string]struct{ seed uint64 }{"plain": {0}, "seed": {goldenSeed}} {
					t.Run(mode, func(t *testing.T) {
						// A consumed block leaves at least a full stripe plus one byte buffered.
						for tail := 65; tail <= 1088; tail++ {
							size := _block + tail
							h := NewSeed(tc.seed)
							h.Write(data[:size])
							if h.blk != 1 || h.len != u64(tail) {
								t.Fatalf("tail %d: blocks=%d, buffered=%d", tail, h.blk, h.len)
							}
							before := *h
							for range 2 {
								if got, want := h.Sum64(), HashSeed(data[:size], tc.seed); got != want {
									t.Fatalf("tail %d: Sum64=%016x, want %016x", tail, got, want)
								}
								if got, want := h.Sum128(), Hash128Seed(data[:size], tc.seed); got != want {
									t.Fatalf("tail %d: Sum128=%x, want %x", tail, got, want)
								}
								if *h != before {
									t.Fatalf("tail %d: Sum changed the Hasher state", tail)
								}
							}
							h.Write(data[size : size+97])
							if got, want := h.Sum64(), HashSeed(data[:size+97], tc.seed); got != want {
								t.Fatalf("tail %d: continued Sum64=%016x, want %016x", tail, got, want)
							}
							if got, want := h.Sum128(), Hash128Seed(data[:size+97], tc.seed); got != want {
								t.Fatalf("tail %d: continued Sum128=%x, want %x", tail, got, want)
							}
						}
					})
				}
			})
		})
	}
}
