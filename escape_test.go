package xxh3

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/zeebo/assert"
)

func TestEscapes(t *testing.T) {
	for _, size := range [...]int{8, 241, 1025} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			tests := map[string]struct {
				fn func()
			}{
				"Hash": {func() {
					var buf [1025]byte
					_ = Hash(buf[:size])
				}},
				"Hash128": {func() {
					var buf [1025]byte
					_ = Hash128(buf[:size])
				}},
				// Construct string views without measuring a byte-to-string copy.
				"HashString": {func() {
					var buf [1025]byte
					_ = HashString(unsafe.String(&buf[0], size))
				}},
				"HashString128": {func() {
					var buf [1025]byte
					_ = HashString128(unsafe.String(&buf[0], size))
				}},
				"HashSeed": {func() {
					var buf [1025]byte
					_ = HashSeed(buf[:size], 1)
				}},
				"Hash128Seed": {func() {
					var buf [1025]byte
					_ = Hash128Seed(buf[:size], 1)
				}},
				"HashStringSeed": {func() {
					var buf [1025]byte
					_ = HashStringSeed(unsafe.String(&buf[0], size), 1)
				}},
				"HashString128Seed": {func() {
					var buf [1025]byte
					_ = HashString128Seed(unsafe.String(&buf[0], size), 1)
				}},
			}
			for name, test := range tests {
				t.Run(name, func(t *testing.T) {
					assert.Equal(t, testing.AllocsPerRun(100, test.fn), 0.0)
				})
			}
		})
	}
}
