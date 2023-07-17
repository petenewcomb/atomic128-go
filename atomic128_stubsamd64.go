//go:build !amd64 || gccgo || appengine
// +build !amd64 gccgo appengine

package atomic128

func compareAndSwapUint128amd64(*[2]uint64, [2]uint64, [2]uint64) bool { panic("not implemented") }
func loadUint128amd64(*[2]uint64) [2]uint64                            { panic("not implemented") }
func storeUint128amd64(*[2]uint64, [2]uint64)                          { panic("not implemented") }
func swapUint128amd64(*[2]uint64, [2]uint64) [2]uint64                 { panic("not implemented") }
func addUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64         { panic("not implemented") }
func andUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64         { panic("not implemented") }
func orUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64          { panic("not implemented") }
func xorUint128amd64(ptr *[2]uint64, incr [2]uint64) [2]uint64         { panic("not implemented") }
