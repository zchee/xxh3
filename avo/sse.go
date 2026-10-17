package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func SSE() {
	// Lay out the prime constant in memory
	primeData := GLOBL("prime_sse", RODATA|NOPTR)
	DATA(0, U32(2654435761))
	DATA(4, U32(2654435761))
	DATA(8, U32(2654435761))
	DATA(12, U32(2654435761))

	initData := GLOBL("init_acc_sse", RODATA|NOPTR)
	for i, v := range [...]uint64{
		3266489917, 11400714785074694791, 14029467366897019727, 1609587929392839161,
		9650029242287828579, 2246822519, 2870177450012600261, 2654435761,
	} {
		DATA(i*8, U64(v))
	}

	oneShot := func(name string, initFromTable bool) {
		TEXT(name, NOSPLIT, "func(acc *[8]uint64, data, key *byte, len uint64)")

		acc := Mem{Base: Load(Param("acc"), GP64())}
		data := Mem{Base: Load(Param("data"), GP64())}
		key := Mem{Base: Load(Param("key"), GP64())}
		skey := Mem{Base: Load(Param("key"), GP64())}
		plen := Load(Param("len"), GP64())
		prime := XMM()
		a := [4]VecVirtual{XMM(), XMM(), XMM(), XMM()}
		raw := [4]VecVirtual{XMM(), XMM(), XMM(), XMM()}

		advance := func(n int) {
			ADDQ(U32(n*64), data.Base)
			SUBQ(U32(n*64), plen)
		}

		Label("load")
		{
			source := acc
			if initFromTable {
				// Keep initialisation independent of the caller's accumulator stores.
				source = initData
			}
			MOVOU(source.Offset(0x00), a[0])
			MOVOU(source.Offset(0x10), a[1])
			MOVOU(source.Offset(0x20), a[2])
			MOVOU(source.Offset(0x30), a[3])
			MOVOU(primeData, prime)
		}

		Label("accum_large")
		{
			CMPQ(plen, U32(1024))
			JLE(LabelRef("accum_init"))

			for i := 0; i < 16; i++ {
				sse2accum(data, a, raw, 64*i, 8*i, key, i == 0)
			}
			advance(16)
			// The scramble is nonlinear, so it must include the permuted raw sums.
			for n, v := range raw {
				PSHUFD(Imm(78), v, v)
				PADDQ(v, a[n])
			}
			sse2scramble(prime, key, a, 8*16)

			JMP(LabelRef("accum_large"))
		}

		Label("accum_init")
		{
			// A partial block must not reuse the preceding block's raw sums.
			for _, v := range raw {
				PXOR(v, v)
			}
		}

		Label("accum")
		{
			CMPQ(plen, Imm(64))
			JLE(LabelRef("finalize"))

			sse2accum(data, a, raw, 0, 0, skey, false)
			advance(1)
			ADDQ(U32(8), skey.Base)

			JMP(LabelRef("accum"))
		}

		Label("finalize")
		{
			CMPQ(plen, Imm(0))
			JE(LabelRef("return"))

			SUBQ(Imm(64), data.Base)
			ADDQ(plen, data.Base)

			sse2accum(data, a, raw, 0, 121, key, false)
		}

		Label("return")
		{
			// Include the final stripe before permuting raw sums into the result.
			for n, v := range raw {
				PSHUFD(Imm(78), v, v)
				PADDQ(v, a[n])
			}
			MOVOU(a[0], acc.Offset(0x00))
			MOVOU(a[1], acc.Offset(0x10))
			MOVOU(a[2], acc.Offset(0x20))
			MOVOU(a[3], acc.Offset(0x30))
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
		prime := XMM()
		a := [4]VecVirtual{XMM(), XMM(), XMM(), XMM()}
		raw := [4]VecVirtual{XMM(), XMM(), XMM(), XMM()}

		Label("load")
		{
			source := acc
			if initFromTable {
				// Keep initialisation independent of the caller's accumulator stores.
				source = initData
			}
			MOVOU(source.Offset(0x00), a[0])
			MOVOU(source.Offset(0x10), a[1])
			MOVOU(source.Offset(0x20), a[2])
			MOVOU(source.Offset(0x30), a[3])
			MOVOU(primeData, prime)
		}

		if multiple {
			TESTQ(blocks, blocks)
			JE(LabelRef("return"))
		}
		Label("accum_block")
		{
			for i := 0; i < 16; i++ {
				if multiple {
					// Fetch one block ahead without entering another Go call.
					PREFETCHT0(data.Offset(64*i + 1024))
				}
				sse2accum(data, a, raw, 64*i, 8*i, key, i == 0)
			}
			// The scramble is nonlinear, so it must include the permuted raw sums.
			for n, v := range raw {
				PSHUFD(Imm(78), v, v)
				PADDQ(v, a[n])
			}
			sse2scramble(prime, key, a, 8*16)
			if multiple {
				ADDQ(U32(1024), data.Base)
				SUBQ(Imm(1), blocks)
				JNE(LabelRef("accum_block"))
			}
		}

		Label("return")
		{
			MOVOU(a[0], acc.Offset(0x00))
			MOVOU(a[1], acc.Offset(0x10))
			MOVOU(a[2], acc.Offset(0x20))
			MOVOU(a[3], acc.Offset(0x30))
			RET()
		}
	}

	oneShot("accumSSE", false)
	block("accumBlockSSE", false, false)
	oneShot("accumSSEInit", true)
	block("accumBlockSSEInit", true, false)
	block("accumBlocksSSE", false, true)
	Generate()
}

func sse2scramble(prime VecVirtual, key Mem, a [4]VecVirtual, koff int) {
	for n, offset := range []int{0x00, 0x10, 0x20, 0x30} {
		x0, x1 := XMM(), XMM()

		MOVOU(a[n], x0)
		PSRLQ(Imm(0x2f), x0)
		PXOR(x0, a[n])
		MOVOU(key.Offset(koff+offset), x0)
		PXOR(x0, a[n])
		PSHUFD(Imm(0xf5), a[n], x1) // 3 3 1 1
		PMULULQ(prime, a[n])
		PMULULQ(prime, x1)
		PSLLQ(Imm(0x20), x1)

		PADDQ(x1, a[n])
	}
}

func sse2accum(data Mem, a, raw [4]VecVirtual, doff, koff int, key Mem, initRaw bool) {
	for n, offset := range []int{0x00, 0x10, 0x20, 0x30} {
		x0, x1, x2 := XMM(), XMM(), XMM()
		if initRaw {
			// Seed raw sums with the first stripe instead of zeroing and adding.
			x0 = raw[n]
		}

		MOVOU(data.Offset(doff+offset), x0)
		MOVOU(key.Offset(koff+offset), x1)
		PXOR(x0, x1)
		PSHUFD(Imm(49), x1, x2)
		PMULULQ(x1, x2)
		// Swapping adjacent 64-bit elements commutes with addition modulo 2^64.
		if !initRaw {
			PADDQ(x0, raw[n])
		}
		PADDQ(x2, a[n])
	}
}
