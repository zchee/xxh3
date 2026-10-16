//go:build arm64 && !go1.27

#include "textflag.h"

// Lane addition is modulo 2^64 and linear: swap(sum(data)) = sum(swap(data)).
// Sum raw data in V21-V24 and swap once per block instead of once per stripe.
// Products alternate between V1-V4 and V11-V14 to shorten dependency chains.
// V0 holds four copies of the scramble prime; V31 holds {0, prime, 0, prime}.

// Widen the low 32-bit products and add them to the 64-bit accumulator.
#define UMLAL_2D(d, n, m) WORD $(0x2EA08000 | ((m) << 16) | ((n) << 5) | (d))
// Widen and accumulate the upper two 32-bit products.
#define UMLAL2_2D(d, n, m) WORD $(0x6EA08000 | ((m) << 16) | ((n) << 5) | (d))
// Initialise an accumulator from the low two widened 32-bit products.
#define UMULL_2D(d, n, m) WORD $(0x2EA0C000 | ((m) << 16) | ((n) << 5) | (d))
// Initialise an accumulator from the upper two widened 32-bit products.
#define UMULL2_2D(d, n, m) WORD $(0x6EA0C000 | ((m) << 16) | ((n) << 5) | (d))
// Multiply each 32-bit element, keeping the low half for the scramble.
#define MUL_4S(d, n, m) WORD $(0x4EA09C00 | ((m) << 16) | ((n) << 5) | (d))
// Extract the low 32 bits of each 64-bit accumulator for widening multiplication.
#define XTN_2S(d, n) WORD $(0x0EA12800 | ((n) << 5) | (d))

// Load a stripe and its key with encodable pair offsets; KR is key or key+8.
#define LOADS(doff, KR, koff) \
	FLDPQ doff(R1), (F16, F17)    \
	FLDPQ doff+32(R1), (F18, F19) \
	FLDPQ koff(KR), (F5, F6)      \
	FLDPQ koff+32(KR), (F7, F8)

// Accumulate unswapped data so the lane swap can be deferred until block end.
#define DSUM \
	VADD V16.D2, V21.D2, V21.D2 \
	VADD V17.D2, V22.D2, V22.D2 \
	VADD V18.D2, V23.D2, V23.D2 \
	VADD V19.D2, V24.D2, V24.D2

// XOR keys, then unzip low and high 32-bit halves of two vectors at once.
// Go lists the second source first; preserving this order preserves lane order.
#define EORUZP(S0, S1, S2, S3) \
	VEOR  S0.B16, V5.B16, V25.B16 \
	VEOR  S1.B16, V6.B16, V26.B16 \
	VEOR  S2.B16, V7.B16, V27.B16 \
	VEOR  S3.B16, V8.B16, V28.B16 \
	VUZP1 V26.S4, V25.S4, V29.S4  \
	VUZP2 V26.S4, V25.S4, V30.S4  \
	VUZP1 V28.S4, V27.S4, V9.S4   \
	VUZP2 V28.S4, V27.S4, V10.S4

// Fuse multiplication with accumulation for all eight hash lanes.
#define MLA(a0, a1, a2, a3) \
	UMLAL_2D(a0, 29, 30)  \
	UMLAL2_2D(a1, 29, 30) \
	UMLAL_2D(a2, 9, 10)   \
	UMLAL2_2D(a3, 9, 10)

// Initialise odd-stripe accumulators from products without zeroing them first.
#define MUL4(a0, a1, a2, a3) \
	UMULL_2D(a0, 29, 30)  \
	UMULL2_2D(a1, 29, 30) \
	UMULL_2D(a2, 9, 10)   \
	UMULL2_2D(a3, 9, 10)

// Even and odd stripes use independent product chains but share the raw sums.
#define STRIPE_EVEN(doff, KR, koff) LOADS(doff, KR, koff); DSUM; EORUZP(V16, V17, V18, V19); MLA(1, 2, 3, 4)
#define STRIPE_ODD(doff, KR, koff) LOADS(doff, KR, koff); DSUM; EORUZP(V16, V17, V18, V19); MLA(11, 12, 13, 14)

// Merge product chains, swap the summed data lanes once, and add them.
#define FOLD(A, B, DS) \
	VADD B.D2, A.D2, A.D2           \
	VEXT $8, DS.B16, DS.B16, DS.B16 \
	VADD DS.D2, A.D2, A.D2

// Scramble: acc ^= acc>>47; acc ^= key; multiply modulo 2^64 by the prime.
// The masked 32-bit multiply supplies the high half; UMLAL adds lo32(acc)*prime.
#define SCR(VA, a, VK) \
	VUSHR $47, VA.D2, V20.D2      \
	VEOR  V20.B16, VA.B16, VA.B16 \
	VEOR  VK.B16, VA.B16, VA.B16  \
	XTN_2S(20, a)                 \
	MUL_4S(a, a, 31)              \
	UMLAL_2D(a, 20, 0)

// func accumNEON(acc *[8]uint64, data *byte, key *byte, len uint64)
TEXT ·accumNEON(SB), NOSPLIT, $0-32
	MOVD acc+0(FP), R0
	MOVD data+8(FP), R1
	MOVD key+16(FP), R2
	MOVD len+24(FP), R3

	// Load accumulators: V1=[acc0,acc1], V2=[acc2,acc3], V3=[acc4,acc5], V4=[acc6,acc7]
	VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]

	CMP $1024, R3
	BLE accum

	// Only full blocks need scramble constants; the tail does not use them.
	MOVW $0x9e3779b1, R4
	VDUP R4, V0.S4
	VSHL $32, V0.D2, V31.D2
	ADD $8, R2, R8 // Odd stripes use aligned offsets from key+8.

accum_large:
	// Initialise raw sums from stripe zero without clearing four vectors.
	FLDPQ (R1), (F21, F22)
	FLDPQ 32(R1), (F23, F24)
	FLDPQ (R2), (F5, F6)
	FLDPQ 32(R2), (F7, F8)
	EORUZP(V21, V22, V23, V24)
	MLA(1, 2, 3, 4)

	// Stripe one initialises the independent odd product chain.
	LOADS(64, R8, 0)
	DSUM
	EORUZP(V16, V17, V18, V19)
	MUL4(11, 12, 13, 14)

	STRIPE_EVEN(128, R2, 16)
	STRIPE_ODD(192, R8, 16)
	STRIPE_EVEN(256, R2, 32)
	STRIPE_ODD(320, R8, 32)
	STRIPE_EVEN(384, R2, 48)
	STRIPE_ODD(448, R8, 48)
	STRIPE_EVEN(512, R2, 64)
	STRIPE_ODD(576, R8, 64)
	STRIPE_EVEN(640, R2, 80)
	STRIPE_ODD(704, R8, 80)
	STRIPE_EVEN(768, R2, 96)
	STRIPE_ODD(832, R8, 96)
	STRIPE_EVEN(896, R2, 112)
	STRIPE_ODD(960, R8, 112)

	// Flush deferred data and odd products before the nonlinear scramble.
	FOLD(V1, V11, V21)
	FOLD(V2, V12, V22)
	FOLD(V3, V13, V23)
	FOLD(V4, V14, V24)
	FLDPQ 128(R2), (F5, F6)
	FLDPQ 160(R2), (F7, F8)
	SCR(V1, 1, V5)
	SCR(V2, 2, V6)
	SCR(V3, 3, V7)
	SCR(V4, 4, V8)

	ADD $1024, R1, R1
	SUB $1024, R3, R3
	CMP $1024, R3
	BGT accum_large

	// Align the short-input entry so its setup is independent of the bulk body size.
	PCALIGN $16
accum:
	// A zero-length defensive call leaves the incoming accumulators unchanged.
	CBZ R3, done

	// Process (length-1)/64 regular stripes, then the overlapping last stripe.
	SUB $1, R3, R6
	LSR $6, R6, R6
	ADD R3, R1, R7
	SUB $64, R7, R7 // Preserve the last-stripe address as R1 advances.
	MOVD R2, R5 // The regular key offset advances by eight bytes per stripe.
	VEOR V21.B16, V21.B16, V21.B16
	VEOR V22.B16, V22.B16, V22.B16
	VEOR V23.B16, V23.B16, V23.B16
	VEOR V24.B16, V24.B16, V24.B16
	CBZ R6, finalize

	// Keep the repeated tail stripe aligned despite changes to the bulk kernel.
	PCALIGN $16
accum_loop:
	// Load the upper pair before post-indexing R1 past this complete stripe.
	FLDPQ 32(R1), (F18, F19)
	FLDPQ.P 64(R1), (F16, F17)
	FLDPQ (R5), (F5, F6)
	FLDPQ 32(R5), (F7, F8)
	ADD $8, R5, R5
	DSUM
	EORUZP(V16, V17, V18, V19)
	MLA(1, 2, 3, 4)
	SUBS $1, R6, R6
	BNE accum_loop

finalize:
	// The overlapping last stripe always uses key offset 121, not the next key.
	ADD $121, R2, R5
	FLDPQ (R7), (F16, F17)
	FLDPQ 32(R7), (F18, F19)
	FLDPQ (R5), (F5, F6)
	FLDPQ 32(R5), (F7, F8)
	DSUM
	EORUZP(V16, V17, V18, V19)
	MLA(1, 2, 3, 4)

	// Fold swapped raw data once, including the overlapping final stripe.
	VEXT $8, V21.B16, V21.B16, V21.B16
	VADD V21.D2, V1.D2, V1.D2
	VEXT $8, V22.B16, V22.B16, V22.B16
	VADD V22.D2, V2.D2, V2.D2
	VEXT $8, V23.B16, V23.B16, V23.B16
	VADD V23.D2, V3.D2, V3.D2
	VEXT $8, V24.B16, V24.B16, V24.B16
	VADD V24.D2, V4.D2, V4.D2

done:
	VST1 [V1.D2, V2.D2, V3.D2, V4.D2], (R0)
	RET

// func accumBlockNEON(acc *[8]uint64, data *byte, key *byte)
// Processes exactly 1024 bytes (16 stripes) and scrambles
TEXT ·accumBlockNEON(SB), NOSPLIT, $0-24
	MOVD acc+0(FP), R0
	MOVD data+8(FP), R1
	MOVD key+16(FP), R2

	VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]

	// Build the prime and its high-half mask once for this block.
	MOVW $0x9e3779b1, R4
	VDUP R4, V0.S4
	VSHL $32, V0.D2, V31.D2
	ADD $8, R2, R8 // Odd stripes use encodable aligned key offsets.

	// Stripe zero initialises the raw sums, retaining incoming accumulators.
	FLDPQ (R1), (F21, F22)
	FLDPQ 32(R1), (F23, F24)
	FLDPQ (R2), (F5, F6)
	FLDPQ 32(R2), (F7, F8)
	EORUZP(V21, V22, V23, V24)
	MLA(1, 2, 3, 4)

	// Stripe one starts a second product chain without zeroing it.
	LOADS(64, R8, 0)
	DSUM
	EORUZP(V16, V17, V18, V19)
	MUL4(11, 12, 13, 14)

	STRIPE_EVEN(128, R2, 16)
	STRIPE_ODD(192, R8, 16)
	STRIPE_EVEN(256, R2, 32)
	STRIPE_ODD(320, R8, 32)
	STRIPE_EVEN(384, R2, 48)
	STRIPE_ODD(448, R8, 48)
	STRIPE_EVEN(512, R2, 64)
	STRIPE_ODD(576, R8, 64)
	STRIPE_EVEN(640, R2, 80)
	STRIPE_ODD(704, R8, 80)
	STRIPE_EVEN(768, R2, 96)
	STRIPE_ODD(832, R8, 96)
	STRIPE_EVEN(896, R2, 112)
	STRIPE_ODD(960, R8, 112)

	// Fold both product chains and swapped raw sums before scrambling.
	FOLD(V1, V11, V21)
	FOLD(V2, V12, V22)
	FOLD(V3, V13, V23)
	FOLD(V4, V14, V24)
	FLDPQ 128(R2), (F5, F6)
	FLDPQ 160(R2), (F7, F8)
	SCR(V1, 1, V5)
	SCR(V2, 2, V6)
	SCR(V3, 3, V7)
	SCR(V4, 4, V8)

	VST1 [V1.D2, V2.D2, V3.D2, V4.D2], (R0)
	RET
