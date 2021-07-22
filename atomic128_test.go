package atomic128_test

import (
	"testing"

	"github.com/CAFxX/atomic128"
)

func TestLoadStore(t *testing.T) {
	n := &atomic128.Uint128{}

	v := atomic128.LoadUint128(n)
	if got, expected := v, [2]uint64{0, 0}; got != expected {
		t.Fatalf("got %v, expected %v", got, expected)
	}

	atomic128.StoreUint128(n, [2]uint64{1, ^uint64(0)})
	v = atomic128.LoadUint128(n)
	if got, expected := v, [2]uint64{1, ^uint64(0)}; got != expected {
		t.Fatalf("got %v, expected %v", got, expected)
	}
}
