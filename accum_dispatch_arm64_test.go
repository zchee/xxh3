//go:build arm64

package xxh3

func withAVX512(cb func())  { cb() }
func withAVX2(cb func())    { cb() }
func withSSE2(cb func())    { cb() }
func withGeneric(cb func()) { cb() }
