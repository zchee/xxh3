package xxh3

import (
	"math/bits"
)

// Hash128Seed returns the 128-bit hash of the byte slice with given seed.
func Hash128Seed(b []byte, seed uint64) Uint128 {
	return hashAny128Seed(*(*str)(ptr(&b)), seed)
}

// HashString128Seed returns the 128-bit hash of the string with given seed.
func HashString128Seed(s string, seed uint64) Uint128 {
	return hashAny128Seed(*(*str)(ptr(&s)), seed)
}

func hashAny128Seed(s str, seed uint64) (acc u128) {
	p, l := s.p, s.l

	switch {
	case l <= 16:
		switch {
		case l > 8: // 9-16
			bitflipl := (key64_032 ^ key64_040) - seed
			bitfliph := (key64_048 ^ key64_056) + seed

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
			seed ^= u64(bits.ReverseBytes32(u32(seed))) << 32
			bitflip := (key64_016 ^ key64_024) + seed
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
			bitflipl := key64_064 ^ key64_072 ^ seed
			bitfliph := key64_080 ^ key64_088 ^ seed
			return u128{Lo: xxh64AvalancheFull(bitflipl), Hi: xxh64AvalancheFull(bitfliph)}
		}

		acc.Hi = uint64(bits.RotateLeft32(bits.ReverseBytes32(uint32(acc.Lo)), 13))
		acc.Lo ^= uint64(key32_000^key32_004) + seed
		acc.Hi ^= uint64(key32_008^key32_012) - seed

		acc.Lo = xxh64AvalancheFull(acc.Lo)
		acc.Hi = xxh64AvalancheFull(acc.Hi)

		return acc

	case l <= 128:
		acc.Lo = u64(l) * prime64_1
		k := key

		if l > 32 {
			if l > 64 {
				if l > 96 {
					acc.Hi += mulFold64(readU64(p, ui(l)-8*8)^(midKey(k, 112, key64_112)+seed), readU64(p, ui(l)-7*8)^(midKey(k, 120, key64_120)-seed))
					acc.Hi ^= readU64(p, 6*8) + readU64(p, 7*8)
					acc.Lo += mulFold64(readU64(p, 6*8)^(midKey(k, 96, key64_096)+seed), readU64(p, 7*8)^(midKey(k, 104, key64_104)-seed))
					acc.Lo ^= readU64(p, ui(l)-8*8) + readU64(p, ui(l)-7*8)

				} // 96

				acc.Hi += mulFold64(readU64(p, ui(l)-6*8)^(midKey(k, 80, key64_080)+seed), readU64(p, ui(l)-5*8)^(midKey(k, 88, key64_088)-seed))
				acc.Hi ^= readU64(p, 4*8) + readU64(p, 5*8)
				acc.Lo += mulFold64(readU64(p, 4*8)^(midKey(k, 64, key64_064)+seed), readU64(p, 5*8)^(midKey(k, 72, key64_072)-seed))
				acc.Lo ^= readU64(p, ui(l)-6*8) + readU64(p, ui(l)-5*8)

			} // 64

			acc.Hi += mulFold64(readU64(p, ui(l)-4*8)^(midKey(k, 48, key64_048)+seed), readU64(p, ui(l)-3*8)^(midKey(k, 56, key64_056)-seed))
			acc.Hi ^= readU64(p, 2*8) + readU64(p, 3*8)
			acc.Lo += mulFold64(readU64(p, 2*8)^(midKey(k, 32, key64_032)+seed), readU64(p, 3*8)^(midKey(k, 40, key64_040)-seed))
			acc.Lo ^= readU64(p, ui(l)-4*8) + readU64(p, ui(l)-3*8)

		} // 32

		acc.Hi += mulFold64(readU64(p, ui(l)-2*8)^(midKey(k, 16, key64_016)+seed), readU64(p, ui(l)-1*8)^(midKey(k, 24, key64_024)-seed))
		acc.Hi ^= readU64(p, 0*8) + readU64(p, 1*8)
		acc.Lo += mulFold64(readU64(p, 0*8)^(midKey(k, 0, key64_000)+seed), readU64(p, 1*8)^(midKey(k, 8, key64_008)-seed))
		acc.Lo ^= readU64(p, ui(l)-2*8) + readU64(p, ui(l)-1*8)

		acc.Hi, acc.Lo = (acc.Lo*prime64_1)+(acc.Hi*prime64_4)+((u64(l)-seed)*prime64_2), acc.Hi+acc.Lo

		acc.Hi = -xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		return acc

	case l <= 240:
		acc.Lo = u64(l) * prime64_1
		k := key

		acc.Hi += mulFold64(readU64(p, 2*8)^(midKey(k, 16, key64_016)+seed), readU64(p, 3*8)^(midKey(k, 24, key64_024)-seed))
		acc.Hi ^= readU64(p, 0*8) + readU64(p, 1*8)
		acc.Lo += mulFold64(readU64(p, 0*8)^(midKey(k, 0, key64_000)+seed), readU64(p, 1*8)^(midKey(k, 8, key64_008)-seed))
		acc.Lo ^= readU64(p, 2*8) + readU64(p, 3*8)

		acc.Hi += mulFold64(readU64(p, 6*8)^(midKey(k, 48, key64_048)+seed), readU64(p, 7*8)^(midKey(k, 56, key64_056)-seed))
		acc.Hi ^= readU64(p, 4*8) + readU64(p, 5*8)
		acc.Lo += mulFold64(readU64(p, 4*8)^(midKey(k, 32, key64_032)+seed), readU64(p, 5*8)^(midKey(k, 40, key64_040)-seed))
		acc.Lo ^= readU64(p, 6*8) + readU64(p, 7*8)

		acc.Hi += mulFold64(readU64(p, 10*8)^(midKey(k, 80, key64_080)+seed), readU64(p, 11*8)^(midKey(k, 88, key64_088)-seed))
		acc.Hi ^= readU64(p, 8*8) + readU64(p, 9*8)
		acc.Lo += mulFold64(readU64(p, 8*8)^(midKey(k, 64, key64_064)+seed), readU64(p, 9*8)^(midKey(k, 72, key64_072)-seed))
		acc.Lo ^= readU64(p, 10*8) + readU64(p, 11*8)

		acc.Hi += mulFold64(readU64(p, 14*8)^(midKey(k, 112, key64_112)+seed), readU64(p, 15*8)^(midKey(k, 120, key64_120)-seed))
		acc.Hi ^= readU64(p, 12*8) + readU64(p, 13*8)
		acc.Lo += mulFold64(readU64(p, 12*8)^(midKey(k, 96, key64_096)+seed), readU64(p, 13*8)^(midKey(k, 104, key64_104)-seed))
		acc.Lo ^= readU64(p, 14*8) + readU64(p, 15*8)

		acc.Hi = xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		// trailing groups after 128
		if l >= 160 {
			acc.Hi += mulFold64(readU64(p, 144)^(readU64(k, 19)+seed), readU64(p, 152)^(readU64(k, 27)-seed))
			acc.Hi ^= readU64(p, 128) + readU64(p, 136)
			acc.Lo += mulFold64(readU64(p, 128)^(readU64(k, 3)+seed), readU64(p, 136)^(readU64(k, 11)-seed))
			acc.Lo ^= readU64(p, 144) + readU64(p, 152)

			if l >= 192 {
				acc.Hi += mulFold64(readU64(p, 176)^(readU64(k, 51)+seed), readU64(p, 184)^(readU64(k, 59)-seed))
				acc.Hi ^= readU64(p, 160) + readU64(p, 168)
				acc.Lo += mulFold64(readU64(p, 160)^(readU64(k, 35)+seed), readU64(p, 168)^(readU64(k, 43)-seed))
				acc.Lo ^= readU64(p, 176) + readU64(p, 184)

				if l >= 224 {
					acc.Hi += mulFold64(readU64(p, 208)^(readU64(k, 83)+seed), readU64(p, 216)^(readU64(k, 91)-seed))
					acc.Hi ^= readU64(p, 192) + readU64(p, 200)
					acc.Lo += mulFold64(readU64(p, 192)^(readU64(k, 67)+seed), readU64(p, 200)^(readU64(k, 75)-seed))
					acc.Lo ^= readU64(p, 208) + readU64(p, 216)
				}
			}
		}

		// last 32 bytes
		{
			seed := 0 - seed
			acc.Hi += mulFold64(readU64(p, ui(l)-32)^(midKey(k, 119, key64_119)+seed), readU64(p, ui(l)-24)^(midKey(k, 127, key64_127)-seed))
			acc.Hi ^= readU64(p, ui(l)-16) + readU64(p, ui(l)-8)
			acc.Lo += mulFold64(readU64(p, ui(l)-16)^(midKey(k, 103, key64_103)+seed), readU64(p, ui(l)-8)^(midKey(k, 111, key64_111)-seed))
			acc.Lo ^= readU64(p, ui(l)-32) + readU64(p, ui(l)-24)
		}

		acc.Hi, acc.Lo = (acc.Lo*prime64_1)+(acc.Hi*prime64_4)+((u64(l)-seed)*prime64_2), acc.Hi+acc.Lo

		acc.Hi = -xxh3Avalanche(acc.Hi)
		acc.Lo = xxh3Avalanche(acc.Lo)

		return acc

	default:
		if seed == 0 {
			return hashLarge128Secret(p, l, key)
		}
		var secret [secretSize]byte
		if hasAVX2 {
			initSecretAVX2(ptr(&secret), seed)
		} else if hasNEON {
			initSecretNEON(ptr(&secret), key, seed)
		} else {
			initSecret(ptr(&secret), seed)
		}
		return hashLarge128Secret(p, l, ptr(&secret))
	}
}

// hashLarge128Secret hashes more than 240 bytes with an already-derived secret.
func hashLarge128Secret(p ptr, l uint, secret ptr) u128 {
	acc := u128{Lo: u64(l) * prime64_1, Hi: ^(u64(l) * prime64_2)}

	var accs [8]u64
	if hasAVX512 && l > avx512Switch {
		accumAVX512Init(&accs, p, secret, u64(l))
	} else if hasAVX2 {
		accumAVX2Init(&accs, p, secret, u64(l))
	} else if hasSSE2 {
		accumSSEInit(&accs, p, secret, u64(l))
	} else if hasNEON {
		accumNEONInit(&accs, p, secret, u64(l))
	} else {
		accs = [8]u64{
			prime32_3, prime64_1, prime64_2, prime64_3,
			prime64_4, prime32_2, prime64_5, prime32_1,
		}
		accumScalar(&accs, p, secret, u64(l))
	}

	const hi_off = 117 - 11

	acc.Lo += mulFold64(accs[0]^readU64(secret, 11), accs[1]^readU64(secret, 19))
	acc.Hi += mulFold64(accs[0]^readU64(secret, 11+hi_off), accs[1]^readU64(secret, 19+hi_off))

	acc.Lo += mulFold64(accs[2]^readU64(secret, 27), accs[3]^readU64(secret, 35))
	acc.Hi += mulFold64(accs[2]^readU64(secret, 27+hi_off), accs[3]^readU64(secret, 35+hi_off))

	acc.Lo += mulFold64(accs[4]^readU64(secret, 43), accs[5]^readU64(secret, 51))
	acc.Hi += mulFold64(accs[4]^readU64(secret, 43+hi_off), accs[5]^readU64(secret, 51+hi_off))

	acc.Lo += mulFold64(accs[6]^readU64(secret, 59), accs[7]^readU64(secret, 67))
	acc.Hi += mulFold64(accs[6]^readU64(secret, 59+hi_off), accs[7]^readU64(secret, 67+hi_off))

	acc.Lo = xxh3Avalanche(acc.Lo)
	acc.Hi = xxh3Avalanche(acc.Hi)

	return acc
}
