package xxh3

import (
	"math/bits"
	"unsafe"
)

// Uint128 is a 128 bit value.
// The actual value can be thought of as u.Hi<<64 | u.Lo.
type Uint128 struct {
	Hi, Lo uint64
}

// Bytes returns the uint128 as an array of bytes in canonical form (big-endian encoded).
func (u Uint128) Bytes() [16]byte {
	return [16]byte{
		byte(u.Hi >> 0x38), byte(u.Hi >> 0x30), byte(u.Hi >> 0x28), byte(u.Hi >> 0x20),
		byte(u.Hi >> 0x18), byte(u.Hi >> 0x10), byte(u.Hi >> 0x08), byte(u.Hi),
		byte(u.Lo >> 0x38), byte(u.Lo >> 0x30), byte(u.Lo >> 0x28), byte(u.Lo >> 0x20),
		byte(u.Lo >> 0x18), byte(u.Lo >> 0x10), byte(u.Lo >> 0x08), byte(u.Lo),
	}
}

type (
	ptr = unsafe.Pointer
	ui  = uintptr

	u8   = uint8
	u32  = uint32
	u64  = uint64
	u128 = Uint128
)

type str struct {
	p ptr
	l uint
}

func readU8(p ptr, o ui) uint8 {
	return *(*uint8)(unsafe.Add(p, o))
}

func readU16(p ptr, o ui) uint16 {
	b := (*[2]byte)(unsafe.Add(p, o))
	return uint16(b[0]) | uint16(b[1])<<8
}

func readU32(p ptr, o ui) uint32 {
	b := (*[4]byte)(unsafe.Add(p, o))
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func readU64(p ptr, o ui) uint64 {
	b := (*[8]byte)(unsafe.Add(p, o))
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func writeU64(p ptr, o ui, v u64) {
	b := (*[8]byte)(unsafe.Add(p, o))
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

const secretSize = 192

func initSecret(secret ptr, seed u64) {
	k := key
	writeU64(secret, 0, readU64(k, 0)+seed)
	writeU64(secret, 8, readU64(k, 8)-seed)
	writeU64(secret, 16, readU64(k, 16)+seed)
	writeU64(secret, 24, readU64(k, 24)-seed)
	writeU64(secret, 32, readU64(k, 32)+seed)
	writeU64(secret, 40, readU64(k, 40)-seed)
	writeU64(secret, 48, readU64(k, 48)+seed)
	writeU64(secret, 56, readU64(k, 56)-seed)
	writeU64(secret, 64, readU64(k, 64)+seed)
	writeU64(secret, 72, readU64(k, 72)-seed)
	writeU64(secret, 80, readU64(k, 80)+seed)
	writeU64(secret, 88, readU64(k, 88)-seed)
	writeU64(secret, 96, readU64(k, 96)+seed)
	writeU64(secret, 104, readU64(k, 104)-seed)
	writeU64(secret, 112, readU64(k, 112)+seed)
	writeU64(secret, 120, readU64(k, 120)-seed)
	writeU64(secret, 128, readU64(k, 128)+seed)
	writeU64(secret, 136, readU64(k, 136)-seed)
	writeU64(secret, 144, readU64(k, 144)+seed)
	writeU64(secret, 152, readU64(k, 152)-seed)
	writeU64(secret, 160, readU64(k, 160)+seed)
	writeU64(secret, 168, readU64(k, 168)-seed)
	writeU64(secret, 176, readU64(k, 176)+seed)
	writeU64(secret, 184, readU64(k, 184)-seed)
}

func xxh64AvalancheSmall(x u64) u64 {
	// x ^= x >> 33                    // x must be < 32 bits
	// x ^= u64(key32_000 ^ key32_004) // caller must do this
	x *= prime64_2
	x ^= x >> 29
	x *= prime64_3
	x ^= x >> 32
	return x
}

func xxhAvalancheSmall(x u64) u64 {
	x ^= x >> 33
	x *= prime64_2
	x ^= x >> 29
	x *= prime64_3
	x ^= x >> 32
	return x
}

func xxh64AvalancheFull(x u64) u64 {
	x ^= x >> 33
	x *= prime64_2
	x ^= x >> 29
	x *= prime64_3
	x ^= x >> 32
	return x
}

func xxh3Avalanche(x u64) u64 {
	x ^= x >> 37
	x *= 0x165667919e3779f9
	x ^= x >> 32
	return x
}

func rrmxmx(h64, len u64) u64 {
	h64 ^= bits.RotateLeft64(h64, 49) ^ bits.RotateLeft64(h64, 24)
	h64 *= 0x9fb21c651e98df25
	h64 ^= (h64 >> 35) + len
	h64 *= 0x9fb21c651e98df25
	h64 ^= (h64 >> 28)
	return h64
}

func mulFold64(x, y u64) u64 {
	hi, lo := bits.Mul64(x, y)
	return hi ^ lo
}
