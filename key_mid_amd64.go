//go:build amd64

package xxh3

// midKey keeps the immediate on amd64 because it takes one instruction,
// while a table read adds a memory operand.
func midKey(k ptr, off uintptr, imm u64) u64 { return imm }
