//go:build amd64 && !gccgo && !appengine
// +build amd64,!gccgo,!appengine

package atomic128

import "github.com/klauspost/cpuid/v2"

func compareAndSwapUint128amd64(*[2]uint64, [2]uint64, [2]uint64) bool
func loadUint128amd64(*[2]uint64) [2]uint64
func storeUint128amd64(*[2]uint64, [2]uint64)
func swapUint128amd64(*[2]uint64, [2]uint64) [2]uint64
func addUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64
func andUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64
func orUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64
func xorUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64

func init() {
	haveNativeAmd64 = cpuid.CPU.Supports(cpuid.CX16)
	if haveNativeAmd64 {
		useNativeAmd64 = true
	}
}
