package atomic128

import (
	"sync"
	"unsafe"
)

type Uint128 struct {
	m sync.Mutex
	d [3]uint64
}

func CompareAndSwapUint128(ptr *Uint128, old, new [2]uint64) bool {
	if compareAndSwapUint128 != nil {
		return compareAndSwapUint128(addr(ptr), old, new)
	}

	ptr.m.Lock()
	v := load(ptr)
	if v != old {
		ptr.m.Unlock()
		return false
	}
	store(ptr, new)
	ptr.m.Unlock()
	return true
}

func LoadUint128(ptr *Uint128) [2]uint64 {
	if loadUint128 != nil {
		return loadUint128(addr(ptr))
	}

	ptr.m.Lock()
	v := load(ptr)
	ptr.m.Unlock()
	return v
}

func StoreUint128(ptr *Uint128, new [2]uint64) {
	if storeUint128 != nil {
		storeUint128(addr(ptr), new)
		return
	}

	ptr.m.Lock()
	store(ptr, new)
	ptr.m.Unlock()
}

func SwapUint128(ptr *Uint128, new [2]uint64) [2]uint64 {
	if swapUint128 != nil {
		return swapUint128(addr(ptr), new)
	}

	ptr.m.Lock()
	old := load(ptr)
	store(ptr, new)
	ptr.m.Unlock()
	return old
}

func AddUint128(ptr *Uint128, incr [2]uint64) [2]uint64 {
	if addUint128 != nil {
		return addUint128(addr(ptr), incr)
	}

	ptr.m.Lock()
	v := load(ptr)
	v[0] += incr[0]
	if v[0] < incr[0] {
		v[1]++
	}
	v[1] += incr[1]
	store(ptr, v)
	ptr.m.Unlock()
	return v
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

var (
	compareAndSwapUint128 func(*[2]uint64, [2]uint64, [2]uint64) bool
	loadUint128           func(*[2]uint64) [2]uint64
	storeUint128          func(*[2]uint64, [2]uint64)
	swapUint128           func(*[2]uint64, [2]uint64) [2]uint64
	addUint128            func(*[2]uint64, [2]uint64) [2]uint64
)
