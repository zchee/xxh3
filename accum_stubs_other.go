//go:build !amd64 && !arm64
// +build !amd64,!arm64

package xxh3

import (
	"unsafe"
)

const (
	hasAVX2   = false
	hasSSE2   = false
	hasAVX512 = false
	hasNEON   = false
)

func initSecretAVX2(secret unsafe.Pointer, seed u64)      { panic("unreachable") }
func initSecretNEON(secret, key unsafe.Pointer, seed u64) { panic("unreachable") }

func accumAVX2(acc *[8]u64, data, key unsafe.Pointer, len u64)     { panic("unreachable") }
func accumSSE(acc *[8]u64, data, key unsafe.Pointer, len u64)      { panic("unreachable") }
func accumBlockAVX2(acc *[8]u64, data, key unsafe.Pointer)         { panic("unreachable") }
func accumBlockSSE(acc *[8]u64, data, key unsafe.Pointer)          { panic("unreachable") }
func accumAVX512(acc *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumBlockAVX512(acc *[8]u64, data, key unsafe.Pointer)       { panic("unreachable") }
func accumNEON(acc *[8]u64, data, key unsafe.Pointer, len u64)     { panic("unreachable") }
func accumNEONInit(acc *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }
func accumBlockNEON(acc *[8]u64, data, key unsafe.Pointer)         { panic("unreachable") }

func accumAVX2From(acc, src *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumAVX512From(acc, src *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }

func accumAVX2Init(acc *[8]u64, data, key unsafe.Pointer, len u64)   { panic("unreachable") }
func accumAVX512Init(acc *[8]u64, data, key unsafe.Pointer, len u64) { panic("unreachable") }
func accumSSEInit(acc *[8]u64, data, key unsafe.Pointer, len u64)    { panic("unreachable") }
func accumBlockAVX2Init(acc *[8]u64, data, key unsafe.Pointer)       { panic("unreachable") }
func accumBlockAVX512Init(acc *[8]u64, data, key unsafe.Pointer)     { panic("unreachable") }
func accumBlockSSEInit(acc *[8]u64, data, key unsafe.Pointer)        { panic("unreachable") }

func accumBlocksAVX2(acc *[8]u64, data, key unsafe.Pointer, n u64)   { panic("unreachable") }
func accumBlocksAVX512(acc *[8]u64, data, key unsafe.Pointer, n u64) { panic("unreachable") }
func accumBlocksSSE(acc *[8]u64, data, key unsafe.Pointer, n u64)    { panic("unreachable") }
func accumBlocksNEON(acc *[8]u64, data, key unsafe.Pointer, n u64)   { panic("unreachable") }
