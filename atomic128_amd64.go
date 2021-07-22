// +build amd64,!gccgo,!appengine

package atomic128

import "github.com/klauspost/cpuid/v2"

func compareAndSwapUint128amd64(*[2]uint64, [2]uint64, [2]uint64) bool
func loadUint128amd64(*[2]uint64) [2]uint64
func storeUint128amd64(*[2]uint64, [2]uint64)
func swapUint128amd64(*[2]uint64, [2]uint64) [2]uint64

func addUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64 {
	for {
		old := loadUint128amd64(ptr)
		new := old
		new[0] += incr[0]
		if new[0] < incr[0] {
			new[1]++
		}
		new[1] += incr[1]
		if compareAndSwapUint128amd64(ptr, old, new) {
			return new
		}
	}
}

func init() {
	if !cpuid.CPU.Supports(cpuid.CX16) {
		return
	}
	compareAndSwapUint128 = compareAndSwapUint128amd64
	loadUint128 = loadUint128amd64
	storeUint128 = storeUint128amd64
	swapUint128 = swapUint128amd64
	addUint128 = addUint128amd64
}
