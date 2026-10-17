package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// AVX512 can be used for very long blocks.
func AVX512() {
	// Lay out the prime constant in memory, copy it so no unpack is needed.
	primeData := GLOBL("prime_avx512", RODATA|NOPTR)
	for i := 0; i < 64; i += 8 {
		DATA(i, U64(2654435761))
	}

	initData := GLOBL("init_acc_avx512", RODATA|NOPTR)
	for i, v := range [...]uint64{
		3266489917, 11400714785074694791, 14029467366897019727, 1609587929392839161,
		9650029242287828579, 2246822519, 2870177450012600261, 2654435761,
	} {
		DATA(i*8, U64(v))
	}

	oneShot := func(name string, initFromTable, separateSource bool) {
		signature := "func(acc *[8]uint64, data, key *byte, len uint64)"
		if separateSource {
			signature = "func(acc, src *[8]uint64, data, key *byte, len uint64)"
		}
		TEXT(name, NOSPLIT, signature)

		acc := Mem{Base: Load(Param("acc"), GP64())}
		data := Mem{Base: Load(Param("data"), GP64())}
		key := Mem{Base: Load(Param("key"), GP64())}
		plen := Load(Param("len"), GP64())
		prime := ZMM()
		// a[0] contains the initial accumulators and completed blocks.
		// a[1:3] sum products; a[3] and raw sum odd and even raw data.
		a := [4]VecVirtual{ZMM(), ZMM(), ZMM(), ZMM()}
		raw := ZMM()

		Label("load")
		{
			source := acc
			if initFromTable {
				// Avoid forwarding narrow Go stores into wide vector loads.
				source = initData
			} else if separateSource {
				source = Mem{Base: Load(Param("src"), GP64())}
			}
			VMOVDQU64(source.Offset(0x00), a[0])
			VMOVDQU64(primeData, prime)
		}
		// Load key at 8 byte offsets in the order we use it.
		var keyReg [17]VecVirtual
		for i := range keyReg {
			keyReg[i] = ZMM()
			VMOVDQU64(key.Offset(i*8), keyReg[i])
		}
		advance := func(n int) {
			ADDQ(U32(n*64), data.Base)
			SUBQ(U32(n*64), plen)
		}

		// Key at offset 121.
		key121 := ZMM()
		VMOVDQU64(key.Offset(121), key121)

		Label("accum_large")
		{
			CMPQ(plen, U32(1024))
			JLE(LabelRef("accum"))

			for i := 0; i < 16; i += 2 {
				avx512accumSplitTwo(data, a, raw, 64*i, keyReg[i], keyReg[i+1], true, i == 0)
			}
			// Permute raw sums once before the nonlinear scramble.
			VPSHUFD(Imm(78), raw, raw)
			VPSHUFD(Imm(78), a[3], a[3])
			VPADDQ(a[0], raw, a[0])
			VPADDQ(a[0], a[1], a[0]) // a0 = a0 + a1
			VPADDQ(a[2], a[3], a[2]) // a2 = a2 + a3
			VPADDQ(a[0], a[2], a[0]) // a0 = a0 + a2
			advance(16)
			avx512scramble(prime, keyReg[16], a[0])

			JMP(LabelRef("accum_large"))
		}

		Label("accum")
		{
			// A partial block must not reuse the preceding block's raw sums.
			VPXORD(raw, raw, raw)
			// Unroll loop so we can use already loaded keys.
			for i := 0; i < 1024/64; i++ {
				CMPQ(plen, Imm(64))
				JLE(LabelRef("finalize"))

				avx512accum(data, a[0], raw, 0, keyReg[i], false)
				advance(1)
			}
		}

		Label("finalize")
		{
			CMPQ(plen, Imm(0))
			JE(LabelRef("return"))

			SUBQ(Imm(64), data.Base)
			ADDQ(plen, data.Base)

			avx512accum(data, a[0], raw, 0, key121, false)
		}

		Label("return")
		{
			// Include the final stripe before permuting raw sums into the result.
			VPSHUFD(Imm(78), raw, raw)
			VPADDQ(a[0], raw, a[0])
			VMOVDQU64(a[0], acc.Offset(0x00))
			VZEROUPPER()
			RET()
		}
	}

	block := func(name string, initFromTable, multiple bool) {
		signature := "func(acc *[8]uint64, data, key *byte)"
		if multiple {
			signature = "func(acc *[8]uint64, data, key *byte, n uint64)"
		}
		TEXT(name, NOSPLIT, signature)

		acc := Mem{Base: Load(Param("acc"), GP64())}
		data := Mem{Base: Load(Param("data"), GP64())}
		key := Mem{Base: Load(Param("key"), GP64())}
		var blocks Register
		if multiple {
			blocks = Load(Param("n"), GP64())
		}
		prime := ZMM()
		// a[0] contains the initial accumulators and completed blocks.
		// a[1:3] sum products; a[3] and raw sum odd and even raw data.
		a := [4]VecVirtual{ZMM(), ZMM(), ZMM(), ZMM()}
		raw := ZMM()

		Label("load")
		{
			source := acc
			if initFromTable {
				// Avoid forwarding narrow Go stores into wide vector loads.
				source = initData
			}
			VMOVDQU64(source.Offset(0x00), a[0])
			VMOVDQU64(primeData, prime)
		}

		var keyReg [17]VecVirtual
		if multiple {
			// Amortise all key loads across the complete bulk write.
			for i := range keyReg {
				keyReg[i] = ZMM()
				VMOVDQU64(key.Offset(i*8), keyReg[i])
			}
			TESTQ(blocks, blocks)
			JE(LabelRef("return"))
		}
		Label("accum_blocks")
		{
			for i := 0; i < 16; i += 2 {
				key1, key2 := keyReg[i], keyReg[i+1]
				if !multiple {
					// Single-block calls load keys just in time.
					key1, key2 = ZMM(), ZMM()
					VMOVDQU64(key.Offset(i*8), key1)
					VMOVDQU64(key.Offset(i*8+8), key2)
				}
				avx512accumSplitTwo(data, a, raw, 64*i, key1, key2, multiple, i == 0)
			}
			// Permute both raw sums before combining products and scrambling.
			VPSHUFD(Imm(78), raw, raw)
			VPSHUFD(Imm(78), a[3], a[3])
			key128 := keyReg[16]
			if !multiple {
				key128 = ZMM()
				VMOVDQU64(key.Offset(16*8), key128)
			}
			VPADDQ(a[0], raw, a[0])
			VPADDQ(a[0], a[1], a[0]) // a0 = a0 + a1
			VPADDQ(a[2], a[3], a[2]) // a2 = a2 + a3
			VPADDQ(a[0], a[2], a[0]) // a0 = a0 + a2
			avx512scramble(prime, key128, a[0])
			if multiple {
				ADDQ(U32(1024), data.Base)
				SUBQ(Imm(1), blocks)
				JNE(LabelRef("accum_blocks"))
			}
		}

		Label("return")
		{
			VMOVDQU64(a[0], acc.Offset(0x00))
			VZEROUPPER()
			RET()
		}
	}

	oneShot("accumAVX512", false, false)
	block("accumBlockAVX512", false, false)
	oneShot("accumAVX512Init", true, false)
	oneShot("accumAVX512From", false, true)
	block("accumBlockAVX512Init", true, false)
	block("accumBlocksAVX512", false, true)
	Generate()
}

func avx512scramble(prime, key, a VecVirtual) {
	y0, y1 := ZMM(), ZMM()

	VPSRLQ(Imm(0x2f), a, y0)
	// 3 way Xor, replacing VPXOR(a, y0, y0), VPXOR(key[koff], y0, y0)
	VPTERNLOGD(U8(0x96), a, key, y0)
	VPMULUDQ(prime, y0, y1)
	VPSHUFD(Imm(0xf5), y0, y0)
	VPMULUDQ(prime, y0, y0)
	VPSLLQ(Imm(0x20), y0, y0)

	VPADDQ(y1, y0, a)
}

func avx512accum(data Mem, a, raw VecVirtual, doff int, key VecVirtual, prefetch bool) {
	y0, y1, y2 := ZMM(), ZMM(), ZMM()
	VMOVDQU64(data.Offset(doff+0x00), y0)
	if prefetch {
		// Prefetch once per cacheline (64 bytes), 8 iterations ahead.
		PREFETCHT0(data.Offset(doff + 1024))
	}
	VPXORD(key, y0, y1)
	VPSHUFD(Imm(49), y1, y2)
	VPMULUDQ(y1, y2, y1)
	// Swapping adjacent 64-bit elements commutes with addition modulo 2^64.
	VPADDQ(a, y1, a)
	VPADDQ(raw, y0, raw)
}

func avx512accumSplitTwo(data Mem, a [4]VecVirtual, raw VecVirtual, doff int, key, key2 VecVirtual, prefetch, initA1 bool) {
	y0, y1, y2 := ZMM(), ZMM(), ZMM()
	z0, z1, z2 := ZMM(), ZMM(), ZMM()
	if initA1 {
		// Seed both product and raw sums directly from the first stripe pair.
		y0 = raw
		y1 = a[1]
		z1 = a[2]
		z0 = a[3]
	}
	VMOVDQU64(data.Offset(doff+0x00), y0)
	VMOVDQU64(data.Offset(doff+0x40), z0)
	if prefetch {
		// Prefetch once per cacheline (64 bytes), 8 iterations ahead.
		PREFETCHT0(data.Offset(doff + 1024))
		PREFETCHT0(data.Offset(doff + 1024 + 64))
	}
	VPXORD(key, y0, y1)
	VPXORD(key2, z0, z1)
	VPSHUFD(Imm(49), y1, y2)
	VPSHUFD(Imm(49), z1, z2)
	VPMULUDQ(y1, y2, y1)
	VPMULUDQ(z1, z2, z1)
	if !initA1 {
		VPADDQ(a[1], y1, a[1])
		VPADDQ(a[2], z1, a[2])
		VPADDQ(a[3], z0, a[3])
		VPADDQ(raw, y0, raw)
	}
}
