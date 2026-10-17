//go:build arm64

package xxh3

import "unsafe"

const (
	hasAVX2   = false
	hasSSE2   = false
	hasAVX512 = false
	hasNEON   = true
)

// initSecretNEON fills secret with the default key adjusted by seed.
//
//go:noescape
func initSecretNEON(secret, key unsafe.Pointer, seed u64)

func initSecretAVX2(secret unsafe.Pointer, seed u64) { panic("unreachable") }

//go:noescape
func accumNEON(acc *[8]u64, data, key unsafe.Pointer, len u64)

// accumNEONInit ignores acc's contents and starts from the initial accumulators.
//
//go:noescape
func accumNEONInit(acc *[8]u64, data, key unsafe.Pointer, len u64)

//go:noescape
func accumBlockNEON(acc *[8]u64, data, key unsafe.Pointer)
func accumAVX2(acc *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumSSE(acc *[8]u64, data, key unsafe.Pointer, len u64)    { panic("unreachable") }
func accumBlockAVX2(acc *[8]u64, data, key unsafe.Pointer)       { panic("unreachable") }
func accumBlockSSE(acc *[8]u64, data, key unsafe.Pointer)        { panic("unreachable") }
func accumAVX512(acc *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }
func accumBlockAVX512(acc *[8]u64, data, key unsafe.Pointer)     { panic("unreachable") }

func accumAVX2From(acc, src *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumAVX512From(acc, src *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }

func accumAVX2Init(acc *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumAVX512Init(acc *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }
func accumSSEInit(acc *[8]u64, data, key unsafe.Pointer, len u64)    { panic("unreachable") }
func accumBlockAVX2Init(acc *[8]u64, data, key unsafe.Pointer)       { panic("unreachable") }
func accumBlockAVX512Init(acc *[8]u64, data, key unsafe.Pointer)     { panic("unreachable") }
func accumBlockSSEInit(acc *[8]u64, data, key unsafe.Pointer)        { panic("unreachable") }

// accumBlocksNEON accumulates and scrambles n complete blocks.
//
//go:noescape
func accumBlocksNEON(acc *[8]u64, data, key unsafe.Pointer, n u64)

func accumBlocksAVX2(acc *[8]u64, data, key unsafe.Pointer, n u64)   { panic("unreachable") }
func accumBlocksAVX512(acc *[8]u64, data, key unsafe.Pointer, n u64) { panic("unreachable") }
func accumBlocksSSE(acc *[8]u64, data, key unsafe.Pointer, n u64)    { panic("unreachable") }
