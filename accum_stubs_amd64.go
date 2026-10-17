package xxh3

import (
	"unsafe"

	cpuid "github.com/klauspost/cpuid/v2"
)

var (
	hasAVX2   = cpuid.CPU.Has(cpuid.AVX2)
	hasSSE2   = cpuid.CPU.Has(cpuid.SSE2) // Always true on amd64
	hasAVX512 = cpuid.CPU.Has(cpuid.AVX512F)
)

const hasNEON = false

// initSecretAVX2 fills secret with the default key adjusted by seed.
//
//go:noescape
func initSecretAVX2(secret unsafe.Pointer, seed u64)

func initSecretNEON(secret, key unsafe.Pointer, seed u64) { panic("unreachable") }

//go:noescape
func accumAVX2(acc *[8]u64, data, key unsafe.Pointer, len u64)

//go:noescape
func accumAVX512(acc *[8]u64, data, key unsafe.Pointer, len u64)

//go:noescape
func accumSSE(acc *[8]u64, data, key unsafe.Pointer, len u64)

//go:noescape
func accumBlockAVX2(acc *[8]u64, data, key unsafe.Pointer)

//go:noescape
func accumBlockAVX512(acc *[8]u64, data, key unsafe.Pointer)

//go:noescape
func accumBlockSSE(acc *[8]u64, data, key unsafe.Pointer)

// accumAVX2Init ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumAVX2Init(acc *[8]u64, data, key unsafe.Pointer, len u64)

// accumAVX512Init ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumAVX512Init(acc *[8]u64, data, key unsafe.Pointer, len u64)

// accumSSEInit ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumSSEInit(acc *[8]u64, data, key unsafe.Pointer, len u64)

// accumBlockAVX2Init ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumBlockAVX2Init(acc *[8]u64, data, key unsafe.Pointer)

// accumBlockAVX512Init ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumBlockAVX512Init(acc *[8]u64, data, key unsafe.Pointer)

// accumBlockSSEInit ignores the contents of acc, starts from the xxh3 initial accumulators and stores the result in acc.
//
//go:noescape
func accumBlockSSEInit(acc *[8]u64, data, key unsafe.Pointer)

func accumNEON(acc *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }
func accumBlockNEON(acc *[8]u64, data, key unsafe.Pointer)     { panic("unreachable") }

// accumBlocksAVX2 accumulates and scrambles n complete blocks.
//
//go:noescape
func accumBlocksAVX2(acc *[8]u64, data, key unsafe.Pointer, n u64)

// accumBlocksAVX512 accumulates and scrambles n complete blocks.
//
//go:noescape
func accumBlocksAVX512(acc *[8]u64, data, key unsafe.Pointer, n u64)

// accumBlocksSSE accumulates and scrambles n complete blocks.
//
//go:noescape
func accumBlocksSSE(acc *[8]u64, data, key unsafe.Pointer, n u64)

func accumBlocksNEON(acc *[8]u64, data, key unsafe.Pointer, n u64) { panic("unreachable") }

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
