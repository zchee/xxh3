package xxh3

import (
	"math/bits"
)

// Hash128 returns the 128-bit hash of the byte slice.
func Hash128(b []byte) Uint128 {
	return hashAny128(*(*str)(ptr(&b)))
}

// HashString128 returns the 128-bit hash of the string slice.
func HashString128(s string) Uint128 {
	return hashAny128(*(*str)(ptr(&s)))
}

func hashAny128(s str) (acc u128) {
	p, l := s.p, s.l

	switch {
	case l <= 16:
		switch {
		case l > 8: // 9-16
			const bitflipl = key64_032 ^ key64_040
			const bitfliph = key64_048 ^ key64_056

			input_lo := readU64(p, 0)
			input_hi := readU64(p, ui(l)-8)

			m128_h, m128_l := bits.Mul64(input_lo^input_hi^bitflipl, prime64_1)

			m128_l += uint64(l-1) << 54
			input_hi ^= bitfliph

			m128_h += input_hi + uint64(uint32(input_hi))*(prime32_2-1)

			m128_l ^= bits.ReverseBytes64(m128_h)

			acc.Hi, acc.Lo = bits.Mul64(m128_l, prime64_2)
			acc.Hi += m128_h * prime64_2

			acc.Lo = xxh3Avalanche(acc.Lo)
			acc.Hi = xxh3Avalanche(acc.Hi)

			return acc

		case l > 3: // 4-8
			const bitflip = key64_016 ^ key64_024

			input_lo := readU32(p, 0)
			input_hi := readU32(p, ui(l)-4)
			input_64 := u64(input_lo) + u64(input_hi)<<32
			keyed := input_64 ^ bitflip

			acc.Hi, acc.Lo = bits.Mul64(keyed, prime64_1+(uint64(l)<<2))

			acc.Hi += acc.Lo << 1
			acc.Lo ^= acc.Hi >> 3

			acc.Lo ^= acc.Lo >> 35
			acc.Lo *= 0x9fb21c651e98df25
			acc.Lo ^= acc.Lo >> 28
			acc.Hi = xxh3Avalanche(acc.Hi)

			return acc

		case l == 3: // 3
			c12 := u64(readU16(p, 0))
			c3 := u64(readU8(p, 2))
			acc.Lo = c12<<16 + c3 + 3<<8

		case l > 1: // 2
			c12 := u64(readU16(p, 0))
			acc.Lo = c12*(1<<24+1)>>8 + 2<<8

		case l == 1: // 1
			c1 := u64(readU8(p, 0))
			acc.Lo = c1*(1<<24+1<<16+1) + 1<<8

		default: // 0
			return u128{0x99aa06d3014798d8, 0x6001c324468d497f}
		}

		acc.Hi = uint64(bits.RotateLeft32(bits.ReverseBytes32(uint32(acc.Lo)), 13))
		acc.Lo ^= uint64(key32_000 ^ key32_004)
		acc.Hi ^= uint64(key32_008 ^ key32_012)

		acc.Lo = xxh64AvalancheSmall(acc.Lo)
		acc.Hi = xxh64AvalancheSmall(acc.Hi)

		return acc

	case l <= 128:
		acc.Lo = u64(l) * prime64_1
		k := key

		if l > 32 {
			if l > 64 {
				if l > 96 {
					acc.Hi += mulFold64(readU64(p, ui(l)-8*8)^midKey(k, 112, key64_112), readU64(p, ui(l)-7*8)^midKey(k, 120, key64_120))
					acc.Hi ^= readU64(p, 6*8) + readU64(p, 7*8)
					acc.Lo += mulFold64(readU64(p, 6*8)^midKey(k, 96, key64_096), readU64(p, 7*8)^midKey(k, 104, key64_104))
					acc.Lo ^= readU64(p, ui(l)-8*8) + readU64(p, ui(l)-7*8)

				} // 96

				acc.Hi += mulFold64(readU64(p, ui(l)-6*8)^midKey(k, 80, key64_080), readU64(p, ui(l)-5*8)^midKey(k, 88, key64_088))
				acc.Hi ^= readU64(p, 4*8) + readU64(p, 5*8)
				acc.Lo += mulFold64(readU64(p, 4*8)^midKey(k, 64, key64_064), readU64(p, 5*8)^midKey(k, 72, key64_072))
				acc.Lo ^= readU64(p, ui(l)-6*8) + readU64(p, ui(l)-5*8)

			} // 64

			acc.Hi += mulFold64(readU64(p, ui(l)-4*8)^midKey(k, 48, key64_048), readU64(p, ui(l)-3*8)^midKey(k, 56, key64_056))
			acc.Hi ^= readU64(p, 2*8) + readU64(p, 3*8)
			acc.Lo += mulFold64(readU64(p, 2*8)^midKey(k, 32, key64_032), readU64(p, 3*8)^midKey(k, 40, key64_040))
			acc.Lo ^= readU64(p, ui(l)-4*8) + readU64(p, ui(l)-3*8)

		} // 32

		acc.Hi += mulFold64(readU64(p, ui(l)-2*8)^midKey(k, 16, key64_016), readU64(p, ui(l)-1*8)^midKey(k, 24, key64_024))
		acc.Hi ^= readU64(p, 0*8) + readU64(p, 1*8)
		acc.Lo += mulFold64(readU64(p, 0*8)^midKey(k, 0, key64_000), readU64(p, 1*8)^midKey(k, 8, key64_008))
		acc.Lo ^= readU64(p, ui(l)-2*8) + readU64(p, ui(l)-1*8)

		acc.Hi, acc.Lo = (acc.Lo*prime64_1)+(acc.Hi*prime64_4)+(u64(l)*prime64_2), acc.Hi+acc.Lo

		acc.Hi = -xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		return acc

	case l <= 240:
		acc.Lo = u64(l) * prime64_1
		k := key

		acc.Hi += mulFold64(readU64(p, 2*8)^midKey(k, 16, key64_016), readU64(p, 3*8)^midKey(k, 24, key64_024))
		acc.Hi ^= readU64(p, 0*8) + readU64(p, 1*8)
		acc.Lo += mulFold64(readU64(p, 0*8)^midKey(k, 0, key64_000), readU64(p, 1*8)^midKey(k, 8, key64_008))
		acc.Lo ^= readU64(p, 2*8) + readU64(p, 3*8)

		acc.Hi += mulFold64(readU64(p, 6*8)^midKey(k, 48, key64_048), readU64(p, 7*8)^midKey(k, 56, key64_056))
		acc.Hi ^= readU64(p, 4*8) + readU64(p, 5*8)
		acc.Lo += mulFold64(readU64(p, 4*8)^midKey(k, 32, key64_032), readU64(p, 5*8)^midKey(k, 40, key64_040))
		acc.Lo ^= readU64(p, 6*8) + readU64(p, 7*8)

		acc.Hi += mulFold64(readU64(p, 10*8)^midKey(k, 80, key64_080), readU64(p, 11*8)^midKey(k, 88, key64_088))
		acc.Hi ^= readU64(p, 8*8) + readU64(p, 9*8)
		acc.Lo += mulFold64(readU64(p, 8*8)^midKey(k, 64, key64_064), readU64(p, 9*8)^midKey(k, 72, key64_072))
		acc.Lo ^= readU64(p, 10*8) + readU64(p, 11*8)

		acc.Hi += mulFold64(readU64(p, 14*8)^midKey(k, 112, key64_112), readU64(p, 15*8)^midKey(k, 120, key64_120))
		acc.Hi ^= readU64(p, 12*8) + readU64(p, 13*8)
		acc.Lo += mulFold64(readU64(p, 12*8)^midKey(k, 96, key64_096), readU64(p, 13*8)^midKey(k, 104, key64_104))
		acc.Lo ^= readU64(p, 14*8) + readU64(p, 15*8)

		// avalanche
		acc.Hi = xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		// trailing groups after 128
		if l >= 160 {
			acc.Hi += mulFold64(readU64(p, 144)^readU64(k, 19), readU64(p, 152)^readU64(k, 27))
			acc.Hi ^= readU64(p, 128) + readU64(p, 136)
			acc.Lo += mulFold64(readU64(p, 128)^readU64(k, 3), readU64(p, 136)^readU64(k, 11))
			acc.Lo ^= readU64(p, 144) + readU64(p, 152)

			if l >= 192 {
				acc.Hi += mulFold64(readU64(p, 176)^readU64(k, 51), readU64(p, 184)^readU64(k, 59))
				acc.Hi ^= readU64(p, 160) + readU64(p, 168)
				acc.Lo += mulFold64(readU64(p, 160)^readU64(k, 35), readU64(p, 168)^readU64(k, 43))
				acc.Lo ^= readU64(p, 176) + readU64(p, 184)

				if l >= 224 {
					acc.Hi += mulFold64(readU64(p, 208)^readU64(k, 83), readU64(p, 216)^readU64(k, 91))
					acc.Hi ^= readU64(p, 192) + readU64(p, 200)
					acc.Lo += mulFold64(readU64(p, 192)^readU64(k, 67), readU64(p, 200)^readU64(k, 75))
					acc.Lo ^= readU64(p, 208) + readU64(p, 216)
				}
			}
		}

		// last 32 bytes
		acc.Hi += mulFold64(readU64(p, ui(l)-32)^midKey(k, 119, key64_119), readU64(p, ui(l)-24)^midKey(k, 127, key64_127))
		acc.Hi ^= readU64(p, ui(l)-16) + readU64(p, ui(l)-8)
		acc.Lo += mulFold64(readU64(p, ui(l)-16)^midKey(k, 103, key64_103), readU64(p, ui(l)-8)^midKey(k, 111, key64_111))
		acc.Lo ^= readU64(p, ui(l)-32) + readU64(p, ui(l)-24)

		acc.Hi, acc.Lo = (acc.Lo*prime64_1)+(acc.Hi*prime64_4)+(u64(l)*prime64_2), acc.Hi+acc.Lo

		acc.Hi = -xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		return acc

	default:
		acc.Lo = u64(l) * prime64_1
		acc.Hi = ^(u64(l) * prime64_2)

		var accs [8]u64
		if hasAVX512 && l > avx512Switch {
			accumAVX512Init(&accs, p, key, u64(l))
		} else if hasAVX2 {
			accumAVX2Init(&accs, p, key, u64(l))
		} else if hasSSE2 {
			accumSSEInit(&accs, p, key, u64(l))
		} else if hasNEON {
			accumNEONInit(&accs, p, key, u64(l))
		} else {
			accs = [8]u64{
				prime32_3, prime64_1, prime64_2, prime64_3,
				prime64_4, prime32_2, prime64_5, prime32_1,
			}
			accumScalar(&accs, p, key, u64(l))
		}

		// merge accs
		k := key
		acc.Lo += mulFold64(accs[0]^midKey(k, 11, key64_011), accs[1]^midKey(k, 19, key64_019))
		acc.Hi += mulFold64(accs[0]^midKey(k, 117, key64_117), accs[1]^midKey(k, 125, key64_125))

		acc.Lo += mulFold64(accs[2]^midKey(k, 27, key64_027), accs[3]^midKey(k, 35, key64_035))
		acc.Hi += mulFold64(accs[2]^midKey(k, 133, key64_133), accs[3]^midKey(k, 141, key64_141))

		acc.Lo += mulFold64(accs[4]^midKey(k, 43, key64_043), accs[5]^midKey(k, 51, key64_051))
		acc.Hi += mulFold64(accs[4]^midKey(k, 149, key64_149), accs[5]^midKey(k, 157, key64_157))

		acc.Lo += mulFold64(accs[6]^midKey(k, 59, key64_059), accs[7]^midKey(k, 67, key64_067))
		acc.Hi += mulFold64(accs[6]^midKey(k, 165, key64_165), accs[7]^midKey(k, 173, key64_173))

		acc.Lo = xxh3Avalanche(acc.Lo)
		acc.Hi = xxh3Avalanche(acc.Hi)

		return acc
	}
}
