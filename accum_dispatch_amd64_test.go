package xxh3

func withOverrides(avx512, avx2, sse2 bool, cb func()) {
	avx512Orig, avx2Orig, sse2Orig := hasAVX512, hasAVX2, hasSSE2
	hasAVX512, hasAVX2, hasSSE2 = avx512, avx2, sse2
	defer func() { hasAVX512, hasAVX2, hasSSE2 = avx512Orig, avx2Orig, sse2Orig }()
	cb()
}

func withAVX512(cb func())  { withOverrides(hasAVX512, false, false, cb) }
func withAVX2(cb func())    { withOverrides(false, hasAVX2, false, cb) }
func withSSE2(cb func())    { withOverrides(false, false, hasSSE2, cb) }
func withGeneric(cb func()) { withOverrides(false, false, false, cb) }
