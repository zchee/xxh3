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

// Share the one-shot body between incoming-state and constant-state entries.
// Distinct labels keep both expansions local to their assembly function.
#define ACCUM_BODY(LARGE, TAIL, LOOP, FINAL, DONE) \
	CMP     $1024, R3;                                                      \
	BLE     TAIL;                                                           \
	/* Only full blocks need scramble constants. */                         \
	MOVW    $0x9e3779b1, R4;                                                \
	VDUP    R4, V0.S4;                                                      \
	VSHL    $32, V0.D2, V31.D2;                                             \
	ADD     $8, R2, R8;                                                     \
	LARGE:                                                                  \
	/* Stripe zero starts raw sums; stripe one starts odd products. */      \
	FLDPQ   (R1), (F21, F22);                                               \
	FLDPQ   32(R1), (F23, F24);                                             \
	FLDPQ   (R2), (F5, F6);                                                 \
	FLDPQ   32(R2), (F7, F8);                                               \
	EORUZP(V21, V22, V23, V24);                                             \
	MLA(1, 2, 3, 4);                                                        \
	LOADS(64, R8, 0);                                                       \
	DSUM;                                                                   \
	EORUZP(V16, V17, V18, V19);                                             \
	MUL4(11, 12, 13, 14);                                                   \
	STRIPE_EVEN(128, R2, 16);                                               \
	STRIPE_ODD(192, R8, 16);                                                \
	STRIPE_EVEN(256, R2, 32);                                               \
	STRIPE_ODD(320, R8, 32);                                                \
	STRIPE_EVEN(384, R2, 48);                                               \
	STRIPE_ODD(448, R8, 48);                                                \
	STRIPE_EVEN(512, R2, 64);                                               \
	STRIPE_ODD(576, R8, 64);                                                \
	STRIPE_EVEN(640, R2, 80);                                               \
	STRIPE_ODD(704, R8, 80);                                                \
	STRIPE_EVEN(768, R2, 96);                                               \
	STRIPE_ODD(832, R8, 96);                                                \
	STRIPE_EVEN(896, R2, 112);                                              \
	STRIPE_ODD(960, R8, 112);                                               \
	/* Flush deferred data and odd products before nonlinear scrambling. */ \
	FOLD(V1, V11, V21);                                                     \
	FOLD(V2, V12, V22);                                                     \
	FOLD(V3, V13, V23);                                                     \
	FOLD(V4, V14, V24);                                                     \
	FLDPQ   128(R2), (F5, F6);                                              \
	FLDPQ   160(R2), (F7, F8);                                              \
	SCR(V1, 1, V5);                                                         \
	SCR(V2, 2, V6);                                                         \
	SCR(V3, 3, V7);                                                         \
	SCR(V4, 4, V8);                                                         \
	ADD     $1024, R1, R1;                                                  \
	SUB     $1024, R3, R3;                                                  \
	CMP     $1024, R3;                                                      \
	BGT     LARGE;                                                          \
	/* Keep short-input setup and the repeated tail stripe aligned. */      \
	PCALIGN $16;                                                            \
	TAIL:                                                                   \
	CBZ     R3, DONE;                                                       \
	/* Preserve the overlapping last-stripe address as R1 advances. */      \
	SUB     $1, R3, R6;                                                     \
	LSR     $6, R6, R6;                                                     \
	ADD     R3, R1, R7;                                                     \
	SUB     $64, R7, R7;                                                    \
	MOVD    R2, R5;                                                         \
	VEOR    V21.B16, V21.B16, V21.B16;                                      \
	VEOR    V22.B16, V22.B16, V22.B16;                                      \
	VEOR    V23.B16, V23.B16, V23.B16;                                      \
	VEOR    V24.B16, V24.B16, V24.B16;                                      \
	CBZ     R6, FINAL;                                                      \
	PCALIGN $16;                                                            \
	LOOP:                                                                   \
	/* Load the upper pair before post-indexing past this stripe. */        \
	FLDPQ   32(R1), (F18, F19);                                             \
	FLDPQ.P 64(R1), (F16, F17);                                             \
	FLDPQ   (R5), (F5, F6);                                                 \
	FLDPQ   32(R5), (F7, F8);                                               \
	ADD     $8, R5, R5;                                                     \
	DSUM;                                                                   \
	EORUZP(V16, V17, V18, V19);                                             \
	MLA(1, 2, 3, 4);                                                        \
	SUBS    $1, R6, R6;                                                     \
	BNE     LOOP;                                                           \
	FINAL:                                                                  \
	/* The last stripe uses key offset 121, not the next regular key. */    \
	ADD     $121, R2, R5;                                                   \
	FLDPQ   (R7), (F16, F17);                                               \
	FLDPQ   32(R7), (F18, F19);                                             \
	FLDPQ   (R5), (F5, F6);                                                 \
	FLDPQ   32(R5), (F7, F8);                                               \
	DSUM;                                                                   \
	EORUZP(V16, V17, V18, V19);                                             \
	MLA(1, 2, 3, 4);                                                        \
	/* Fold swapped raw data once, including the last stripe. */            \
	VEXT    $8, V21.B16, V21.B16, V21.B16;                                  \
	VADD    V21.D2, V1.D2, V1.D2;                                           \
	VEXT    $8, V22.B16, V22.B16, V22.B16;                                  \
	VADD    V22.D2, V2.D2, V2.D2;                                           \
	VEXT    $8, V23.B16, V23.B16, V23.B16;                                  \
	VADD    V23.D2, V3.D2, V3.D2;                                           \
	VEXT    $8, V24.B16, V24.B16, V24.B16;                                  \
	VADD    V24.D2, V4.D2, V4.D2;                                           \
	DONE:                                                                   \
	VST1    [V1.D2, V2.D2, V3.D2, V4.D2], (R0);                             \
	RET

// func accumNEON(acc *[8]uint64, data *byte, key *byte, len uint64)
TEXT ·accumNEON(SB), NOSPLIT, $0-32
	MOVD acc+0(FP), R0
	MOVD data+8(FP), R1
	MOVD key+16(FP), R2
	MOVD len+24(FP), R3
	VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]
	ACCUM_BODY(accum_large, accum, accum_loop, finalize, done)

// Initial accumulator values are read directly instead of copied through the stack.
DATA initialAccumulators<>+0(SB)/8, $0x00000000c2b2ae3d
DATA initialAccumulators<>+8(SB)/8, $0x9e3779b185ebca87
DATA initialAccumulators<>+16(SB)/8, $0xc2b2ae3d27d4eb4f
DATA initialAccumulators<>+24(SB)/8, $0x165667b19e3779f9
DATA initialAccumulators<>+32(SB)/8, $0x85ebca77c2b2ae63
DATA initialAccumulators<>+40(SB)/8, $0x0000000085ebca77
DATA initialAccumulators<>+48(SB)/8, $0x27d4eb2f165667c5
DATA initialAccumulators<>+56(SB)/8, $0x000000009e3779b1
GLOBL initialAccumulators<>(SB), RODATA, $64

// func accumNEONInit(acc *[8]uint64, data *byte, key *byte, len uint64)
TEXT ·accumNEONInit(SB), NOSPLIT, $0-32
	MOVD acc+0(FP), R0
	MOVD data+8(FP), R1
	MOVD key+16(FP), R2
	MOVD len+24(FP), R3
	MOVD $initialAccumulators<>(SB), R9
	VLD1 (R9), [V1.D2, V2.D2, V3.D2, V4.D2]
	ACCUM_BODY(init_accum_large, init_accum, init_accum_loop, init_finalize, init_done)

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
	ADD  $8, R2, R8         // Odd stripes use encodable aligned key offsets.

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

// func accumBlocksNEON(acc *[8]uint64, data *byte, key *byte, n uint64)
// Keep accumulators and scramble constants live across n complete blocks.
TEXT ·accumBlocksNEON(SB), NOSPLIT, $0-32
	MOVD acc+0(FP), R0
	MOVD data+8(FP), R1
	MOVD key+16(FP), R2
	MOVD n+24(FP), R3
	CBZ  R3, blocks_done                    // Zero blocks must not read input or change the state.
	VLD1 (R0), [V1.D2, V2.D2, V3.D2, V4.D2]
	MOVW $0x9e3779b1, R4
	VDUP R4, V0.S4
	VSHL $32, V0.D2, V31.D2
	ADD  $8, R2, R8                         // Odd-stripe key offsets stay pair-load encodable.

blocks_loop:
	// Start raw sums and the odd product chain afresh after every scramble.
	FLDPQ (R1), (F21, F22)
	FLDPQ 32(R1), (F23, F24)
	FLDPQ (R2), (F5, F6)
	FLDPQ 32(R2), (F7, F8)
	EORUZP(V21, V22, V23, V24)
	MLA(1, 2, 3, 4)
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

	// Each complete block ends with the same fold and nonlinear scramble.
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
	ADD   $1024, R1, R1
	SUBS  $1, R3, R3
	BNE   blocks_loop
	VST1  [V1.D2, V2.D2, V3.D2, V4.D2], (R0)

blocks_done:
	RET

// Fill the secret in ascending order with alternating additions and subtractions.
TEXT ·initSecretNEON(SB), NOSPLIT, $0-24
	MOVD secret+0(FP), R0
	MOVD key+8(FP), R1
	MOVD seed+16(FP), R2
	NEG  R2, R3
	VDUP R2, V0.D2
	VMOV R3, V0.D[1]

	VLD1.P 64(R1), [V1.D2, V2.D2, V3.D2, V4.D2]
	VADD   V0.D2, V1.D2, V1.D2
	VADD   V0.D2, V2.D2, V2.D2
	VADD   V0.D2, V3.D2, V3.D2
	VADD   V0.D2, V4.D2, V4.D2
	VST1.P [V1.D2, V2.D2, V3.D2, V4.D2], 64(R0)

	VLD1.P 64(R1), [V1.D2, V2.D2, V3.D2, V4.D2]
	VADD   V0.D2, V1.D2, V1.D2
	VADD   V0.D2, V2.D2, V2.D2
	VADD   V0.D2, V3.D2, V3.D2
	VADD   V0.D2, V4.D2, V4.D2
	VST1.P [V1.D2, V2.D2, V3.D2, V4.D2], 64(R0)

	VLD1 (R1), [V1.D2, V2.D2, V3.D2, V4.D2]
	VADD V0.D2, V1.D2, V1.D2
	VADD V0.D2, V2.D2, V2.D2
	VADD V0.D2, V3.D2, V3.D2
	VADD V0.D2, V4.D2, V4.D2
	VST1 [V1.D2, V2.D2, V3.D2, V4.D2], (R0)
	RET
