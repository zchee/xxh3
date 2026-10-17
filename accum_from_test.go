package xxh3

import "testing"

func TestAccumFromParity(t *testing.T) {
	data := goldenData(goldenMaxLength)
	tests := map[string]struct {
		available bool
		accum     func(*[8]u64, *[8]u64, ptr, ptr, u64)
	}{
		"avx2":   {hasAVX2, accumAVX2From},
		"avx512": {hasAVX512, accumAVX512From},
	}
	for name, tt := range tests {
		if !tt.available {
			continue
		}
		t.Run(name, func(t *testing.T) {
			for mode, tc := range map[string]struct{ seed uint64 }{"plain": {0}, "seed": {goldenSeed}} {
				t.Run(mode, func(t *testing.T) {
					var secret [secretSize]byte
					k := key
					if tc.seed != 0 {
						initSecret(ptr(&secret), tc.seed)
						k = ptr(&secret)
					}
					for name, input := range map[string]struct{ offset int }{
						"aligned": {0}, "offset-1": {1}, "offset-7": {7}, "offset-31": {31}, "offset-63": {63},
					} {
						t.Run(name, func(t *testing.T) {
							for tail := 65; tail <= 1088; tail++ {
								src := [8]u64{1, 3, 5, 7, 11, 13, 17, 19}
								before := src
								want := src
								p := ptr(&data[input.offset])
								accumScalar(&want, p, k, u64(tail))
								got := [8]u64{23, 29, 31, 37, 41, 43, 47, 53}
								tt.accum(&got, &src, p, k, u64(tail))
								if got != want || src != before {
									t.Fatalf("tail %d offset %d: got %x, want %x, source changed=%t", tail, input.offset, got, want, src != before)
								}
							}
						})
					}
				})
			}
		})
	}
}
