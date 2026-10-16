//go:build !amd64

package xxh3

import "unsafe"

// midKey reads the table because materializing a 64-bit immediate on arm64
// takes four instructions, making a table load cheaper. The byte-OR body is
// duplicated because a nested inlined call leaves an inline-mark NOP per use on arm64.
func midKey(k ptr, off uintptr, imm u64) u64 {
	b := (*[8]byte)(unsafe.Add(k, off))
	return u64(b[0]) | u64(b[1])<<8 | u64(b[2])<<16 | u64(b[3])<<24 |
		u64(b[4])<<32 | u64(b[5])<<40 | u64(b[6])<<48 | u64(b[7])<<56
}
