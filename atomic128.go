// Package atomic128 implements atomic operations on 128 bit values.
// When possible (e.g. on amd64 processors that support CMPXCHG16B), it automatically uses
// native CPU features to implement the operations; otherwise it falls back to an approach
// based on mutexes.
package atomic128

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

var (
	useNativeAmd64  bool
	haveNativeAmd64 bool
)

func HasNative() bool {
	return haveNativeAmd64
}

func UsingNative() bool {
	return useNativeAmd64
}

func EnableNative() {
	if haveNativeAmd64 {
		useNativeAmd64 = true
	}
}

func DisableNative() {
	useNativeAmd64 = false
}

// Uint128 is an opaque container for an atomic uint128.
// Uint128 must not be copied.
// The zero value must be assumed invalid until Store has been called, because
// of the initialization requirements of atomic.Value in the fallback case.
type Uint128 struct {
	// d is unused in the fallback code path and is length 3 so that addr()
	// can use whichever pair of elements are 128-bit aligned; it is placed first
	// so addr() can rely on Go's 64-bit struct alignment guarantee: see
	// https://go101.org/article/memory-layout.html and, specifically,
	// https://pkg.go.dev/sync/atomic#pkg-note-BUG.
	d  [3]uint64
	av atomic.Value
}

// CompareAndSwapUint128 performs a 128-bit atomic CAS on ptr.
// If the memory pointed to by ptr contains the value old, it is set to
// the value new, and true is returned. Otherwise the memory pointed to
// by ptr is unchanged, and false is returned.
// In the old and new values the first of the two elements is the low-order bits.
func CompareAndSwapUint128(ptr *Uint128, old, new [2]uint64) bool {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return compareAndSwapUint128amd64(addr(ptr), old, new)
	}
	return ptr.av.CompareAndSwap(old, new)
}

// LoadUint128 atomically loads the 128 bit value pointed to by ptr.
// In the returned value the first of the two elements is the low-order bits.
func LoadUint128(ptr *Uint128) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return loadUint128amd64(addr(ptr))
	}
	return ptr.av.Load().([2]uint64)
}

// StoreUint128 atomically stores the new value in the 128 bit value pointed to by ptr.
// In the new value the first of the two elements is the low-order bits.
func StoreUint128(ptr *Uint128, new [2]uint64) {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		storeUint128amd64(addr(ptr), new)
		return
	}
	ptr.av.Store(new)
}

// SwapUint128 atomically stores the new value with the 128 bit value pointed to by ptr,
// and it returns the 128 bit value that was previously pointed to by ptr.
// In the new and returned values the first of the two elements is the low-order bits.
func SwapUint128(ptr *Uint128, new [2]uint64) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return swapUint128amd64(addr(ptr), new)
	}
	return ptr.av.Swap(new).([2]uint64)
}

// AddUint128 atomically adds the incr value to the 128 bit value pointed to by ptr,
// and it returns the resulting 128 bit value.
// In the incr and returned values the first of the two elements is the low-order bits.
func AddUint128(ptr *Uint128, incr [2]uint64) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return addUint128amd64(addr(ptr), incr)
	}

	var new [2]uint64
	for {
		old := LoadUint128(ptr)
		new = old
		new[0] += incr[0]
		if new[0] < incr[0] {
			new[1]++
		}
		new[1] += incr[1]
		if CompareAndSwapUint128(ptr, old, new) {
			break
		}
	}
	return new
}

// AndUint128 atomically performs a bitwise AND of the op value to the 128 bit value pointed to by ptr,
// and it returns the resulting 128 bit value.
// In the op and returned values the first of the two elements is the low-order bits.
func AndUint128(ptr *Uint128, op [2]uint64) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return andUint128amd64(addr(ptr), op)
	}

	var new [2]uint64
	for {
		old := LoadUint128(ptr)
		new = old
		new[0] &= op[0]
		new[1] &= op[1]
		if CompareAndSwapUint128(ptr, old, new) {
			break
		}
	}
	return new
}

// OrUint128 atomically performs a bitwise OR of the op value to the 128 bit value pointed to by ptr,
// and it returns the resulting 128 bit value.
// In the op and returned values the first of the two elements is the low-order bits.
func OrUint128(ptr *Uint128, op [2]uint64) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return orUint128amd64(addr(ptr), op)
	}

	var new [2]uint64
	for {
		old := LoadUint128(ptr)
		new = old
		new[0] |= op[0]
		new[1] |= op[1]
		if CompareAndSwapUint128(ptr, old, new) {
			break
		}
	}
	return new
}

// XorUint128 atomically performs a bitwise XOR of the op value to the 128 bit value pointed to by ptr,
// and it returns the resulting 128 bit value.
// In the op and returned values the first of the two elements is the low-order bits.
func XorUint128(ptr *Uint128, op [2]uint64) [2]uint64 {
	if runtime.GOARCH == "amd64" && useNativeAmd64 {
		return xorUint128amd64(addr(ptr), op)
	}

	var new [2]uint64
	for {
		old := LoadUint128(ptr)
		new = old
		new[0] ^= op[0]
		new[1] ^= op[1]
		if CompareAndSwapUint128(ptr, old, new) {
			break
		}
	}
	return new
}

func addr(ptr *Uint128) *[2]uint64 {
	if (uintptr)((unsafe.Pointer)(&ptr.d[0]))%16 == 0 {
		return (*[2]uint64)((unsafe.Pointer)(&ptr.d[0]))
	}
	return (*[2]uint64)((unsafe.Pointer)(&ptr.d[1]))
}

func load(ptr *Uint128) [2]uint64 {
	return [2]uint64{ptr.d[0], ptr.d[1]}
}

func store(ptr *Uint128, v [2]uint64) {
	ptr.d[0], ptr.d[1] = v[0], v[1]
}
