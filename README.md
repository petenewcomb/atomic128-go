# atomic128-go

[![GoDoc](https://godoc.org/github.com/petenewcomb/atomic128-go?status.svg)](https://godoc.org/github.com/petenewcomb/atomic128-go)
[![Build Status](https://github.com/petenewcomb/atomic128-go/actions/workflows/build.yml/badge.svg)](https://github.com/petenewcomb/atomic128-go/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/petenewcomb/atomic128-go/branch/master/graph/badge.svg?token=03A5UVYW3K)](https://codecov.io/gh/petenewcomb/atomic128-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/petenewcomb/atomic128-go)](https://goreportcard.com/report/github.com/petenewcomb/atomic128-go)

128-bit atomic operations for Golang, using [CMPXCHG16B](http://www.felixcloutier.com/x86/CMPXCHG8B:CMPXCHG16B.html)
when available. 

Based on [github.com/CAFxX/atomic128](https://github.com/CAFxX/atomic128), which is in turn based on [github.com/tmthrgd/atomic128](https://github.com/tmthrgd/atomic128).  This implementation replaces
the mutex-based fallback with a lock-free one based on [`atomic.Value`](https://pkg.go.dev/sync/atomic#Value).

## Performance

As compared to the mutex-based fallback, the one based on `atomic.Value` is faster for `Load`, `Store`, and
`CompareAndSwap` (CAS) operations, but much slower for `Add`, `And`, `Or`, and `Xor` operations. There is a hidden cost,
however, as the all `atomic.Value` operations other than `Load` perform allocations that add to garbage collection
overhead. This may be a reasonable trade-off for some latency-sensitive lock-free data structure use cases that depend
heavily on CAS operations when run with high levels of parallelism.

12-CPU benchmarks (performed on a 12-core machine) show performance under contention:

```
petenewcomb/atomic128-go$ go test -run '^$' -bench=. -benchtime=.1s -count=10 | \
    sed 's|/fallback|/cpu=12/mode=atomic.Value|;s|/native|/cpu=12/mode=native|' >/tmp/12cpu.petenewcomb

CAFxX/atomic128$ go test -run '^$' -bench=. -benchtime=.1s -benchmem -count=10 | \
    sed 's|/fallback|/cpu=12/mode=mutex|;s|/native|/cpu=12/mode=native|' >/tmp/12cpu.CAFxX

$ head -6 /tmp/12cpu.petenewcomb 
goos: linux
goarch: amd64
pkg: github.com/petenewcomb/atomic128-go
cpu: 13th Gen Intel(R) Core(TM) i5-1345U
BenchmarkLoad/cpu=12/mode=native-12              3256340                36.96 ns/op            0 B/op          0 allocs/op
BenchmarkLoad/cpu=12/mode=native-12              2977168                40.05 ns/op            0 B/op          0 allocs/op

$ head -6 /tmp/12cpu.CAFxX 
goos: linux
goarch: amd64
pkg: github.com/CAFxX/atomic128
cpu: 13th Gen Intel(R) Core(TM) i5-1345U
BenchmarkLoad/cpu=12/mode=native-12              3255603                36.74 ns/op            0 B/op          0 allocs/op
BenchmarkLoad/cpu=12/mode=native-12              3260756                38.41 ns/op            0 B/op          0 allocs/op

$ benchstat -filter '.unit:(sec/op OR allocs/op)' -table /cpu -col /mode /tmp/12cpu.CAFxX /tmp/12cpu.petenewcomb
/cpu: 12
         │      native      │                   mutex                    │               atomic.Value                │
         │      sec/op      │     sec/op      vs base                    │    sec/op      vs base                    │
Load-12    38.4300n ±  4% ¹   99.5750n ±  5%  +159.11% (p=0.000 n=20+10)   0.3157n ±  1%   -99.18% (p=0.000 n=20+10)
Store-12    111.45n ± 10% ¹    101.00n ±  2%    -9.38% (p=0.000 n=20+10)    26.34n ± 37%   -76.36% (p=0.000 n=20+10)
Swap-12     111.15n ±  1% ¹     94.67n ±  7%   -14.82% (p=0.000 n=20+10)    50.13n ±  5%   -54.90% (p=0.000 n=20+10)
Add-12      139.70n ± 22% ¹     99.60n ± 13%   -28.70% (p=0.000 n=20+10)   448.70n ±  6%  +221.19% (p=0.000 n=20+10)
And-12       43.53n ± 10% ¹     95.89n ± 10%  +120.25% (p=0.000 n=20+10)   263.95n ± 16%  +506.29% (p=0.000 n=20+10)
Or-12        50.15n ± 21% ¹    102.80n ±  5%  +105.01% (p=0.000 n=20+10)   421.50n ±  5%  +740.56% (p=0.000 n=20+10)
Xor-12       171.3n ± 16% ¹     114.1n ± 11%   -33.42% (p=0.001 n=20+10)    553.0n ±  7%  +222.85% (p=0.000 n=20+10)
CAS-12       56.02n ± 10% ¹    132.05n ± 14%  +135.72% (p=0.000 n=20+10)    64.19n ± 11%   +14.58% (p=0.000 n=20+10)
geomean      78.16n             104.4n         +33.53%                      72.20n          -7.62%
¹ benchmarks vary in pkg

         │    native    │                 mutex                  │            atomic.Value            │
         │  allocs/op   │ allocs/op   vs base                    │  allocs/op   vs base               │
Load-12    0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   0.000 ±  0%  ~ (p=1.000 n=20+10) ²
Store-12   0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ±  0%  ? (p=0.000 n=20+10)
Swap-12    0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ±  0%  ? (p=0.000 n=20+10)
Add-12     0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   6.000 ± 17%  ? (p=0.000 n=20+10)
And-12     0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   3.000 ±  0%  ? (p=0.000 n=20+10)
Or-12      0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   3.000 ±  0%  ? (p=0.000 n=20+10)
Xor-12     0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   6.000 ±  0%  ? (p=0.000 n=20+10)
CAS-12     0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ±  0%  ? (p=0.000 n=20+10)
geomean               ³               +0.00%                   ³                ?                   ³
¹ benchmarks vary in pkg
² all samples are equal
³ summaries must be >0 to compute geomean
```

Single-CPU benchmarks on the same machine to show performance without contention:

```
petenewcomb/atomic128-go$ go test -run '^$' -bench=. -benchtime=.1s -count=10 -cpu 1 | \
    sed 's|/fallback|/cpu=1/mode=atomic.Value|;s|/native|/cpu=1/mode=native|' >/tmp/1cpu.petenewcomb

CAFxX/atomic128$ go test -run '^$' -bench=. -benchtime=.1s -benchmem -count=10 -cpu 1 | \
    sed 's|/fallback|/cpu=1/mode=mutex|;s|/native|/cpu=1/mode=native|' >/tmp/1cpu.CAFxX

$ benchstat -filter '.unit:(sec/op OR allocs/op)' -table /cpu -col /mode /tmp/1cpu.CAFxX /tmp/1cpu.petenewcomb
/cpu: 1
        │     native     │                  mutex                   │               atomic.Value               │
        │     sec/op     │    sec/op      vs base                   │    sec/op     vs base                    │
Load      12.855n ± 0% ¹   11.030n ±  0%  -14.20% (p=0.000 n=20+10)   1.737n ±  0%   -86.49% (p=0.000 n=20+10)
Store      11.22n ± 0% ¹    11.22n ±  0%        ~ (p=0.751 n=20+10)   18.05n ±  1%   +60.92% (p=0.000 n=20+10)
Swap       16.65n ± 0% ¹    11.09n ±  0%  -33.41% (p=0.000 n=20+10)   20.13n ±  0%   +20.86% (p=0.000 n=20+10)
Add        17.04n ± 9% ¹    15.35n ±  0%   -9.94% (p=0.000 n=20+10)   34.40n ±  3%  +101.79% (p=0.000 n=20+10)
And        17.20n ± 6% ¹    11.06n ±  0%  -35.68% (p=0.000 n=20+10)   32.80n ±  0%   +90.75% (p=0.000 n=20+10)
Or         17.16n ± 6% ¹    11.08n ±  0%  -35.46% (p=0.000 n=20+10)   32.76n ± 17%   +90.91% (p=0.000 n=20+10)
Xor        18.32n ± 8% ¹    11.05n ±  0%  -39.68% (p=0.000 n=20+10)   39.54n ±  5%  +115.83% (p=0.000 n=20+10)
CAS        14.45n ± 6% ¹    14.13n ± 12%        ~ (p=0.178 n=20+10)   29.44n ±  4%  +103.70% (p=0.000 n=20+10)
geomean    15.42n           11.90n        -22.81%                     20.15n         +30.65%
¹ benchmarks vary in pkg

        │    native    │                 mutex                  │           atomic.Value            │
        │  allocs/op   │ allocs/op   vs base                    │ allocs/op   vs base               │
Load      0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   0.000 ± 0%  ~ (p=1.000 n=20+10) ²
Store     0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
Swap      0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
Add       0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
And       0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
Or        0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
Xor       0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
CAS       0.000 ± 0% ¹   0.000 ± 0%       ~ (p=1.000 n=20+10) ²   1.000 ± 0%  ? (p=0.000 n=20+10)
geomean              ³               +0.00%                   ³               ?                   ³
¹ benchmarks vary in pkg
² all samples are equal
³ summaries must be >0 to compute geomean
```

## TODO

- Add ARM/aarch64 assembly version
- Add shift/rotate operations
