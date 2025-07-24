# atomic128-go

[![GoDoc](https://godoc.org/github.com/petenewcomb/atomic128-go?status.svg)](https://godoc.org/github.com/petenewcomb/atomic128-go)
[![Build Status](https://github.com/petenewcomb/atomic128-go/actions/workflows/build.yml/badge.svg)](https://github.com/petenewcomb/atomic128-go/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/petenewcomb/atomic128-go/branch/master/graph/badge.svg?token=03A5UVYW3K)](https://codecov.io/gh/petenewcomb/atomic128-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/petenewcomb/atomic128-go)](https://goreportcard.com/report/github.com/petenewcomb/atomic128-go)

128-bit atomic operations for Golang, using [CMPXCHG16B](http://www.felixcloutier.com/x86/CMPXCHG8B:CMPXCHG16B.html)
when available. 

Based on [github.com/CAFxX/atomic128](https://github.com/CAFxX/atomic128), which is in turn based on [github.com/tmthrgd/atomic128](https://github.com/tmthrgd/atomic128).  This implementation replaces
the mutex-based fallback with a lock-free one based on [`atomic.Value`](https://pkg.go.dev/sync/atomic#Value).

This implementation also avoids [use of the BP register on amd64](https://go.dev/doc/asm#amd64), making it safe
to use with Go's [sampling profiler](https://go.dev/blog/pprof) and [execution tracer](https://go.dev/blog/execution-traces-2024).
Using the BP register causes them to occasionally panic.

## Performance

As compared to the mutex-based fallback, the one based on `atomic.Value` is faster for `Load`, `Store`, and
`CompareAndSwap` (CAS) operations, but much slower for `Add`, `And`, `Or`, and `Xor` operations. There is a hidden cost,
however, as the all `atomic.Value` operations other than `Load` perform allocations that add to garbage collection
overhead. This may be a reasonable trade-off for some latency-sensitive lock-free data structure use cases that depend
heavily on CAS operations when run with high levels of parallelism.

Avoiding the BP register on amd64 also comes at a cost for the native path, increasing the execution time for `Add`,
`And`, `Or`, and `Xor` operations there as well.  There is no performance penalty for avoiding BP in `Load`, `Store`,
`Swap`, and CAS operations.

12-CPU benchmarks (performed on a 12-core machine) show performance under contention:

```
petenewcomb/atomic128-go$ go test -run '^$' -bench=. -benchtime=.1s -count=20 | \
    sed 's|/fallback|/cpu=12/mode=atomic.Value|;s|/native|/cpu=12/mode=native|' >/tmp/12cpu.petenewcomb

CAFxX/atomic128$ go test -run '^$' -bench=. -benchtime=.1s -count=20 | \
    sed 's|/fallback|/cpu=12/mode=sync.Mutex|;s|/native|/cpu=12/mode=native|' >/tmp/12cpu.CAFxX

$ head -6 /tmp/12cpu.petenewcomb 
goos: linux
goarch: amd64
pkg: github.com/petenewcomb/atomic128-go
cpu: 13th Gen Intel(R) Core(TM) i5-1345U
BenchmarkLoad/cpu=12/mode=native-12              2977820                40.28 ns/op            0 B/op          0 allocs/op
BenchmarkLoad/cpu=12/mode=native-12              3270016                36.57 ns/op            0 B/op          0 allocs/op

$ head -6 /tmp/12cpu.CAFxX 
goos: linux
goarch: amd64
pkg: github.com/CAFxX/atomic128
cpu: 13th Gen Intel(R) Core(TM) i5-1345U
BenchmarkLoad/cpu=12/mode=native-12              3254995                36.77 ns/op
BenchmarkLoad/cpu=12/mode=native-12              3254300                36.79 ns/op

$ benchstat -filter '.unit:(sec/op OR allocs/op)' -table /cpu -col .file,/mode /tmp/12cpu.CAFxX /tmp/12cpu.petenewcomb
/cpu: 12
         │                    /tmp/12cpu.CAFxX                    │                             /tmp/12cpu.petenewcomb                              │
         │    native     │               sync.Mutex               │                 native                 │              atomic.Value              │
         │    sec/op     │    sec/op      vs base                 │     sec/op      vs base                │    sec/op      vs base                 │
Load-12    36.7750n ± 0%   99.6850n ± 4%  +171.07% (p=0.000 n=20)   36.5800n ±  0%   -0.53% (p=0.000 n=20)   0.3250n ±  1%   -99.12% (p=0.000 n=20)
Store-12    111.05n ± 0%    101.75n ± 1%    -8.37% (p=0.000 n=20)    120.85n ±  9%        ~ (p=0.472 n=20)    42.91n ± 13%   -61.36% (p=0.000 n=20)
Swap-12     111.30n ± 0%     94.27n ± 3%   -15.31% (p=0.000 n=20)    109.15n ±  5%        ~ (p=0.167 n=20)    50.60n ±  2%   -54.53% (p=0.000 n=20)
Add-12       136.9n ± 1%     105.3n ± 3%   -23.05% (p=0.000 n=20)     165.5n ±  4%  +20.85% (p=0.000 n=20)    393.2n ±  7%  +187.25% (p=0.000 n=20)
And-12       40.90n ± 2%    105.50n ± 4%  +157.98% (p=0.000 n=20)     60.53n ±  1%  +48.01% (p=0.000 n=20)   355.25n ±  2%  +768.69% (p=0.000 n=20)
Or-12        40.92n ± 2%    103.90n ± 2%  +153.91% (p=0.000 n=20)     54.36n ± 11%  +32.84% (p=0.000 n=20)   383.05n ± 14%  +836.09% (p=0.000 n=20)
Xor-12       103.8n ± 1%     105.8n ± 2%         ~ (p=0.588 n=20)     132.3n ± 23%  +27.44% (p=0.000 n=20)    563.1n ±  2%  +442.27% (p=0.000 n=20)
CAS-12       43.72n ± 2%    120.10n ± 2%  +174.73% (p=0.000 n=20)     48.58n ±  2%  +11.12% (p=0.000 n=20)    55.64n ±  4%   +27.28% (p=0.000 n=20)
geomean      68.29n          104.3n        +52.78%                    80.07n        +17.25%                   76.59n         +12.15%

         │             /tmp/12cpu.petenewcomb             │
         │    native    │          atomic.Value           │
         │  allocs/op   │  allocs/op   vs base            │
Load-12    0.000 ± 0%     0.000 ±  0%  ~ (p=1.000 n=20) ¹
Store-12   0.000 ± 0%     1.000 ±  0%  ? (p=0.000 n=20)
Swap-12    0.000 ± 0%     1.000 ±  0%  ? (p=0.000 n=20)
Add-12     0.000 ± 0%     6.000 ± 17%  ? (p=0.000 n=20)
And-12     0.000 ± 0%     3.000 ±  0%  ? (p=0.000 n=20)
Or-12      0.000 ± 0%     3.000 ±  0%  ? (p=0.000 n=20)
Xor-12     0.000 ± 0%     6.000 ± 17%  ? (p=0.000 n=20)
CAS-12     0.000 ± 0%     1.000 ±  0%  ? (p=0.000 n=20)
geomean               ²                ?                ²
¹ all samples are equal
² summaries must be >0 to compute geomean
```

Single-CPU benchmarks on the same machine to show performance without contention:

```
petenewcomb/atomic128-go$ go test -run '^$' -bench=. -benchtime=.1s -count=20 -cpu=1 | \
    sed 's|/fallback|/cpu=1/mode=atomic.Value|;s|/native|/cpu=1/mode=native|' >/tmp/1cpu.petenewcomb

CAFxX/atomic128$ go test -run '^$' -bench=. -benchtime=.1s -count=20 -cpu=1 | \
    sed 's|/fallback|/cpu=1/mode=sync.Mutex|;s|/native|/cpu=1/mode=native|' >/tmp/1cpu.CAFxX

$ benchstat -filter '.unit:(sec/op OR allocs/op)' -table /cpu -col .file,/mode /tmp/1cpu.CAFxX /tmp/1cpu.petenewcomb
/cpu: 1
        │                   /tmp/1cpu.CAFxX                   │                            /tmp/1cpu.petenewcomb                            │
        │    native    │              sync.Mutex              │                native                │             atomic.Value             │
        │    sec/op    │    sec/op     vs base                │    sec/op     vs base                │   sec/op     vs base                 │
Load      12.800n ± 0%   10.970n ± 0%  -14.30% (p=0.000 n=20)   12.820n ± 0%        ~ (p=0.187 n=20)   1.730n ± 0%   -86.48% (p=0.000 n=20)
Store      11.16n ± 0%    11.42n ± 0%   +2.33% (p=0.000 n=20)    11.15n ± 0%        ~ (p=0.069 n=20)   17.80n ± 1%   +59.54% (p=0.000 n=20)
Swap       16.59n ± 0%    11.04n ± 0%  -33.45% (p=0.000 n=20)    18.94n ± 0%  +14.17% (p=0.000 n=20)   19.92n ± 0%   +20.04% (p=0.000 n=20)
Add        16.90n ± 0%    15.28n ± 0%   -9.62% (p=0.000 n=20)    18.60n ± 0%  +10.03% (p=0.000 n=20)   34.10n ± 1%  +101.80% (p=0.000 n=20)
And        16.97n ± 0%    11.03n ± 0%  -35.03% (p=0.000 n=20)    16.91n ± 0%   -0.35% (p=0.000 n=20)   32.71n ± 0%   +92.78% (p=0.000 n=20)
Or         17.00n ± 0%    11.01n ± 0%  -35.24% (p=0.000 n=20)    18.40n ± 0%   +8.24% (p=0.000 n=20)   32.61n ± 0%   +91.85% (p=0.000 n=20)
Xor        16.96n ± 0%    10.99n ± 0%  -35.17% (p=0.000 n=20)    18.39n ± 0%   +8.46% (p=0.000 n=20)   32.65n ± 0%   +92.54% (p=0.000 n=20)
CAS        13.44n ± 0%    13.37n ± 0%   -0.48% (p=0.000 n=20)    13.27n ± 0%   -1.23% (p=0.000 n=20)   24.35n ± 0%   +81.21% (p=0.000 n=20)
geomean    15.05n         11.80n       -21.57%                   15.77n        +4.77%                  19.10n        +26.94%

        │             /tmp/1cpu.petenewcomb             │
        │    native    │          atomic.Value          │
        │  allocs/op   │ allocs/op   vs base            │
Load      0.000 ± 0%     0.000 ± 0%  ~ (p=1.000 n=20) ¹
Store     0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
Swap      0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
Add       0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
And       0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
Or        0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
Xor       0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
CAS       0.000 ± 0%     1.000 ± 0%  ? (p=0.000 n=20)
geomean              ²               ?                ²
¹ all samples are equal
² summaries must be >0 to compute geomean
```

## TODO

- Add ARM/aarch64 assembly version
- Add shift/rotate operations
