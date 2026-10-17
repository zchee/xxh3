package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func Secret() {
	TEXT("initSecretAVX2", NOSPLIT, "func(secret *byte, seed uint64)")
	secret := Mem{Base: Load(Param("secret"), GP64())}
	key := Mem{Base: GP64()}
	MOVQ(Mem{Symbol: Symbol{Name: "·key"}, Base: StaticBase}, key.Base)
	seed := Load(Param("seed"), GP64())

	// Repeat the alternating seed pair in both halves of the vector.
	delta := YMM()
	VMOVQ(seed, delta.AsX())
	NEGQ(seed)
	VPINSRQ(U8(1), seed, delta.AsX(), delta.AsX())
	VINSERTI128(U8(1), delta.AsX(), delta, delta)
	value := YMM()
	for offset := 0; offset < 192; offset += 32 {
		VPADDQ(key.Offset(offset), delta, value)
		VMOVDQU(value, secret.Offset(offset))
	}
	VZEROUPPER()
	RET()
	Generate()
}
