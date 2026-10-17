# XXH3
[![GoDoc](https://godoc.org/github.com/zeebo/xxh3?status.svg)](https://godoc.org/github.com/zeebo/xxh3)
[![Sourcegraph](https://sourcegraph.com/github.com/zeebo/xxh3/-/badge.svg)](https://sourcegraph.com/github.com/zeebo/xxh3?badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/zeebo/xxh3)](https://goreportcard.com/report/github.com/zeebo/xxh3)

This package is a port of the [xxh3](https://github.com/Cyan4973/xxHash) library to Go.

Upstream has fixed the output as of v0.8.0, and this package matches that.

---

# Benchmarks

The AMD Ryzen 9 9950X and Intel i7-8850H measurements below predate this branch's changes; that hardware was not re-measured here.

In the benchmark suite, Hasher levels `avx512`, `avx2`, and `sse2` and
Fixed suffixes `-AVX512`, `-AVX2`, and `-SSE2` force the named ISA.
Hasher `go` and unsuffixed Fixed rows force the generic path (scalar on
amd64; NEON remains enabled on arm64). Hasher `auto` and Fixed `-AUTO`
rows use normal, unforced dispatch.

## AMD Ryzen 9 9950X

### Small Sizes

| Bytes      | 64-bit                       | 64-bit Seed                  | 128-bit                      | 128-bit Seed                 |
|------------|------------------------------|------------------------------|------------------------------|------------------------------|
| ` 0 `      | ` 1.16 ns `                  | ` 1.72 ns `                  | ` 1.31 ns `                  | ` 2.13 ns `                  |
| ` 1-3 `    | ` 1.58 ns (0.63-1.90 GB/s) ` | ` 1.79 ns (0.56-1.68 GB/s) ` | ` 1.92 ns (0.52-1.56 GB/s) ` | ` 2.32 ns (0.43-1.29 GB/s) ` |
| ` 4-8 `    | ` 1.56 ns (2.56-5.13 GB/s) ` | ` 1.80 ns (2.22-4.44 GB/s) ` | ` 1.81 ns (2.21-4.42 GB/s) ` | ` 2.15 ns (1.86-3.72 GB/s) ` |
| ` 9-16 `   | ` 1.38 ns (6.52-11.6 GB/s) ` | ` 1.71 ns (5.26-9.36 GB/s) ` | ` 2.25 ns (4.00-7.11 GB/s) ` | ` 2.50 ns (3.60-6.40 GB/s) ` |
| ` 17-32 `  | ` 1.72 ns (9.88-18.6 GB/s) ` | ` 2.07 ns (8.21-15.5 GB/s) ` | ` 2.69 ns (6.32-11.9 GB/s) ` | ` 3.05 ns (5.57-10.5 GB/s) ` |
| ` 33-64 `  | ` 2.39 ns (13.8-26.8 GB/s) ` | ` 2.76 ns (12.0-23.2 GB/s) ` | ` 3.47 ns (9.51-18.4 GB/s) ` | ` 3.87 ns (8.53-16.5 GB/s) ` |
| ` 65-96 `  | ` 3.16 ns (20.6-30.4 GB/s) ` | ` 3.63 ns (17.9-26.4 GB/s) ` | ` 4.46 ns (14.6-21.5 GB/s) ` | ` 4.95 ns (13.1-19.4 GB/s) ` |
| ` 97-128 ` | ` 3.89 ns (24.9-32.9 GB/s) ` | ` 4.41 ns (22.0-29.0 GB/s) ` | ` 5.22 ns (18.6-24.5 GB/s) ` | ` 5.75 ns (16.9-22.3 GB/s) ` |

### Large Sizes

64-bit default is shown. The 128-bit variant adds only a small finalization
(within a few % here). A non-zero seed re-derives the 192-byte secret on every
call, so seeded hashing is markedly slower at these sizes (e.g. 512 B SSE2:
20.3 vs 11.2 ns/op) and only amortizes to <1% by ~100KB.

| Bytes     | Rate                       | SSE2 Rate                  | AVX2 Rate                  | AVX512 Rate                 |
|-----------|----------------------------|----------------------------|----------------------------|-----------------------------|
| ` 129 `   | ` 4.52 ns/op (28.5 GB/s) ` |                            |                            |                             |
| ` 240 `   | ` 7.39 ns/op (32.5 GB/s) ` |                            |                            |                             |
| ` 241 `   | ` 13.6 ns/op (17.7 GB/s) ` | ` 7.95 ns/op (30.3 GB/s) ` | ` 11.3 ns/op (21.4 GB/s) ` |                             |
| ` 512 `   | ` 23.1 ns/op (22.2 GB/s) ` | ` 11.2 ns/op (45.7 GB/s) ` | ` 14.1 ns/op (36.4 GB/s) ` |                             |
| ` 1024 `  | ` 42.1 ns/op (24.3 GB/s) ` | ` 20.7 ns/op (49.5 GB/s) ` | ` 20.9 ns/op (49.0 GB/s) ` | ` 23.9 ns/op (42.9 GB/s) `  |
| ` 100KB ` | ` 4.03 us/op (25.4 GB/s) ` | ` 2.10 us/op (48.7 GB/s) ` | ` 1.07 us/op (96.0 GB/s) ` | ` 685 ns/op (149 GB/s) `    |

## Intel i7-8850H @ 2.60GHz

### Small Sizes

| Bytes      | Rate                                   |
|------------|----------------------------------------|
| ` 0 `      | ` 0.74 ns/op `                         |
| ` 1-3 `    | ` 4.19 ns/op (0.24 GB/s - 0.71 GB/s) ` |
| ` 4-8 `    | ` 4.16 ns/op (0.97 GB/s - 1.98 GB/s) ` |
| ` 9-16 `   | ` 4.46 ns/op (2.02 GB/s - 3.58 GB/s) ` |
| ` 17-32 `  | ` 6.22 ns/op (2.76 GB/s - 5.15 GB/s) ` |
| ` 33-64 `  | ` 8.00 ns/op (4.13 GB/s - 8.13 GB/s) ` |
| ` 65-96 `  | ` 11.0 ns/op (5.91 GB/s - 8.84 GB/s) ` |
| ` 97-128 ` | ` 12.8 ns/op (7.68 GB/s - 10.0 GB/s) ` |

### Large Sizes

| Bytes     | Rate                       | SSE2 Rate                  | AVX2 Rate                  |
|-----------|----------------------------|----------------------------|----------------------------|
| ` 129 `   | ` 13.6 ns/op (9.45 GB/s) ` |                            |                            |
| ` 240 `   | ` 23.8 ns/op (10.1 GB/s) ` |                            |                            |
| ` 241 `   | ` 40.5 ns/op (5.97 GB/s) ` | ` 23.3 ns/op (10.4 GB/s) ` | ` 20.1 ns/op (12.0 GB/s) ` |
| ` 512 `   | ` 69.8 ns/op (7.34 GB/s) ` | ` 30.4 ns/op (16.9 GB/s) ` | ` 24.7 ns/op (20.7 GB/s) ` |
| ` 1024 `  | ` 132  ns/op (7.77 GB/s) ` | ` 48.9 ns/op (20.9 GB/s) ` | ` 37.7 ns/op (27.2 GB/s) ` |
| ` 100KB ` | ` 13.0 us/op (7.88 GB/s) ` | ` 4.05 us/op (25.3 GB/s) ` | ` 2.31 us/op (44.3 GB/s) ` |

## Intel Xeon Platinum 8481C (linux/amd64, Sapphire Rapids VM)

Since the previous benchmark refresh, constant-offset mid-size groups
and long-merge key loads reduce one-shot work; vector secret derivation,
deferred accumulator initialization, separate-source tail kernels and
NEON constant-initialized entries reduce selected seeded/streaming costs.
This comparison uses cumulative baseline `0a183a1`, not the earlier
pre-optimization baseline.

Measured 2026-10-17 06:42:30 UTC, go1.27.2 linux/amd64;
10 alternating samples per leaf, cpu=1, 300 ms per leaf.
The process was pinned to CPU 2. GOEXPERIMENT was empty.
Small rows use the production scalar path. Above 240 B, `-AUTO` uses
unforced dispatch (AVX2 through 512 B, AVX512 at the listed larger sizes).
Hasher rows use the unforced `auto` level.

### Small Sizes

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 0 ` | ` 2.942 ns ` | ` 3.036 ns ` | ` 3.170 ns ` | ` 3.826 ns ` |
| ` 1-3 ` | ` 2.868-3.348 ns (0.34-1.05 GB/s) ` | ` 3.011-3.246 ns (0.33-0.99 GB/s) ` | ` 3.414-3.520 ns (0.29-0.86 GB/s) ` | ` 4.027-4.212 ns (0.25-0.73 GB/s) ` |
| ` 4-8 ` | ` 2.694-2.696 ns (1.48-2.97 GB/s) ` | ` 3.186-3.188 ns (1.26-2.51 GB/s) ` | ` 3.420-3.421 ns (1.17-2.34 GB/s) ` | ` 3.858-3.859 ns (1.04-2.07 GB/s) ` |
| ` 9-16 ` | ` 2.554-2.564 ns (3.52-6.24 GB/s) ` | ` 2.753-2.755 ns (3.27-5.81 GB/s) ` | ` 3.945-3.947 ns (2.28-4.05 GB/s) ` | ` 4.299-4.302 ns (2.09-3.72 GB/s) ` |
| ` 17-32 ` | ` 3.105-3.106 ns (5.47-10.31 GB/s) ` | ` 3.354-3.355 ns (5.07-9.54 GB/s) ` | ` 4.628 ns (3.67-6.91 GB/s) ` | ` 5.034 ns (3.38-6.36 GB/s) ` |
| ` 33-64 ` | ` 4.394 ns (7.51-14.57 GB/s) ` | ` 4.712-4.715 ns (7.00-13.57 GB/s) ` | ` 6.429-6.604 ns (5.00-9.95 GB/s) ` | ` 7.056-7.221 ns (4.57-9.07 GB/s) ` |
| ` 65-96 ` | ` 5.784-5.977 ns (10.88-16.60 GB/s) ` | ` 6.232-6.327 ns (10.27-15.40 GB/s) ` | ` 8.199-8.386 ns (7.75-11.71 GB/s) ` | ` 9.240-9.426 ns (6.90-10.39 GB/s) ` |
| ` 97-128 ` | ` 7.370-7.412 ns (13.09-17.37 GB/s) ` | ` 7.703-7.798 ns (12.44-16.62 GB/s) ` | ` 9.852-9.964 ns (9.74-12.99 GB/s) ` | ` 10.960-11.040 ns (8.79-11.68 GB/s) ` |

Ranges span measured endpoints (plus 2 B in the 1-3 B row), not every intervening size.

### Large Sizes

These are production-dispatch results, not refreshed forced-ISA columns.
A nonzero seed derives the 192-byte secret on each one-shot call above
240 B; Hasher retains its derived secret between writes.

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 129 ` | ` 8.672 ns (14.88 GB/s) ` | ` 9.429 ns (13.68 GB/s) ` | ` 12.640 ns (10.21 GB/s) ` | ` 14.575 ns (8.85 GB/s) ` |
| ` 240 ` | ` 12.430 ns (19.31 GB/s) ` | ` 14.520 ns (16.53 GB/s) ` | ` 17.650 ns (13.60 GB/s) ` | ` 20.935 ns (11.46 GB/s) ` |
| ` 241 ` | ` 11.995 ns (20.09 GB/s) ` | ` 29.560 ns (8.15 GB/s) ` | ` 15.170 ns (15.89 GB/s) ` | ` 30.645 ns (7.86 GB/s) ` |
| ` 512 ` | ` 16.725 ns (30.61 GB/s) ` | ` 32.790 ns (15.61 GB/s) ` | ` 21.660 ns (23.64 GB/s) ` | ` 34.060 ns (15.03 GB/s) ` |
| ` 1024 ` | ` 24.240 ns (42.24 GB/s) ` | ` 45.485 ns (22.51 GB/s) ` | ` 29.705 ns (34.47 GB/s) ` | ` 46.305 ns (22.11 GB/s) ` |
| ` 8192 ` | ` 142.450 ns (57.51 GB/s) ` | ` 166.850 ns (49.10 GB/s) ` | ` 145.750 ns (56.21 GB/s) ` | ` 166.200 ns (49.29 GB/s) ` |
| ` 102400 ` | ` 1.684 us (60.81 GB/s) ` | ` 1.712 us (59.80 GB/s) ` | ` 1.690 us (60.61 GB/s) ` | ` 1.713 us (59.78 GB/s) ` |
| ` 1024000 ` | ` 16.680 us (61.39 GB/s) ` | ` 16.702 us (61.31 GB/s) ` | ` 16.684 us (61.37 GB/s) ` | ` 16.700 us (61.32 GB/s) ` |
| ` 10240000 ` | ` 316.802 us (32.32 GB/s) ` | ` 316.862 us (32.32 GB/s) ` | ` 317.097 us (32.29 GB/s) ` | ` 317.080 us (32.29 GB/s) ` |
| ` 102400000 ` | ` 6.314 ms (16.22 GB/s) ` | ` 6.341 ms (16.15 GB/s) ` | ` 6.366 ms (16.08 GB/s) ` | ` 6.352 ms (16.12 GB/s) ` |

### Streaming

Each operation resets a reused Hasher, writes the complete input, and
computes Sum64 or Sum128. Seed setup is outside the timed loop.

| Bytes | Hasher64 | Hasher64 Seed | Hasher128 | Hasher128 Seed |
|---|---|---|---|---|
| ` 16 ` | ` 8.697 ns (1.84 GB/s) ` | ` 9.266 ns (1.73 GB/s) ` | ` 9.617 ns (1.66 GB/s) ` | ` 10.540 ns (1.52 GB/s) ` |
| ` 64 ` | ` 10.750 ns (5.95 GB/s) ` | ` 12.125 ns (5.28 GB/s) ` | ` 12.315 ns (5.20 GB/s) ` | ` 13.670 ns (4.68 GB/s) ` |
| ` 256 ` | ` 26.570 ns (9.63 GB/s) ` | ` 26.080 ns (9.82 GB/s) ` | ` 28.650 ns (8.94 GB/s) ` | ` 29.485 ns (8.68 GB/s) ` |
| ` 1024 ` | ` 51.705 ns (19.80 GB/s) ` | ` 48.685 ns (21.03 GB/s) ` | ` 52.900 ns (19.36 GB/s) ` | ` 55.180 ns (18.56 GB/s) ` |
| ` 4096 ` | ` 108.600 ns (37.72 GB/s) ` | ` 97.590 ns (41.97 GB/s) ` | ` 106.750 ns (38.37 GB/s) ` | ` 106.100 ns (38.61 GB/s) ` |
| ` 65536 ` | ` 1.116 us (58.72 GB/s) ` | ` 1.110 us (59.07 GB/s) ` | ` 1.120 us (58.51 GB/s) ` | ` 1.118 us (58.62 GB/s) ` |
| ` 1048576 ` | ` 17.114 us (61.27 GB/s) ` | ` 17.122 us (61.24 GB/s) ` | ` 17.126 us (61.23 GB/s) ` | ` 17.128 us (61.22 GB/s) ` |
| ` 16777216 ` | ` 519.428 us (32.30 GB/s) ` | ` 519.591 us (32.29 GB/s) ` | ` 519.400 us (32.30 GB/s) ` | ` 519.934 us (32.27 GB/s) ` |

### Cumulative Comparison

Time geomeans against cumulative baseline `0a183a1` are descriptive;
there is no aggregate p-value. Faster/slower counts use p < 0.05 per row.

| Class | Rows | Time change | Faster / slower |
|---|---:|---:|---:|
| One-shot 0-16 B | 32 | +1.00% | 8 / 22 |
| One-shot 17-128 B | 32 | -2.28% | 30 / 1 |
| One-shot 129-240 B | 8 | -6.63% | 8 / 0 |
| One-shot 241-1024 B | 12 | -1.27% | 8 / 2 |
| One-shot >1024 B | 20 | +0.12% | 0 / 3 |
| Hasher | 52 | -0.32% | 21 / 8 |
| Chunked Hasher | 6 | +0.20% | 0 / 3 |

The measured candidate is `a114709`, inherited by final source tip
`71552bd` through production-symbol identity up to two type-metadata
data relocations; this is not a separate timing of that final tip.

On Xeon, seeded Hasher64 at 16/64 B regresses 9.59%/6.84%
and Hasher128 13.61%/6.59% (median of the base plus three
benchmark-loop placements). The same M3 Max rows improve
9.29%/3.95%/4.78%/7.18%, respectively. Bisection locates the losses at Reset
initialization (`6ecb84e`) and tail-copy removal (`a114709`). These losses
do not come from extra executed work: inlining is unchanged and the
64-bit loop falls from 30 to 21 instructions. An instruction-identical
callee reorder improves three
seeded rows by 4.95-8.84%, confirming layout sensitivity. Restoring the
Reset stores recovers small rows but slows 256 B Hasher64/plain 3.95%
and seeded Hasher64/Hasher128 1.6%/2.2%, so it was not adopted.
Fixed64/2 B default/seed (+4.58%/+7.35%), Fixed64/3 B default
(+4.40%), and Fixed128/0 B seed (+6.51%) remain unresolved layout-class
rows; unseeded Fixed128/0 B changes sign across placements. Most
Fixed64/2 B loss predates `2bd1ec4`. Tables retain the unmodified
cumulative candidate medians, not substituted focused-control results.

## Apple M3 Max (darwin/arm64)

Measured 2026-10-17 15:55:46 JST, go1.27.2 darwin/arm64;
10 alternating samples per leaf, cpu=1, 300 ms per leaf.
GOEXPERIMENT=simd,runtimesecret,mapsplitgroup; header load average 3.44.
Small rows use the production scalar path. Above 240 B, `-AUTO` uses
unforced dispatch (NEON). Hasher rows use the unforced `auto` level.

### Small Sizes

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 0 ` | ` 1.626 ns ` | ` 2.164 ns ` | ` 1.618 ns ` | ` 2.176 ns ` |
| ` 1-3 ` | ` 1.618-1.895 ns (0.53-1.85 GB/s) ` | ` 1.680-1.888 ns (0.53-1.79 GB/s) ` | ` 1.880-2.154 ns (0.46-1.60 GB/s) ` | ` 2.070-2.251 ns (0.44-1.45 GB/s) ` |
| ` 4-8 ` | ` 1.514-1.531 ns (2.64-5.23 GB/s) ` | ` 1.644-1.670 ns (2.40-4.86 GB/s) ` | ` 1.862-1.865 ns (2.15-4.29 GB/s) ` | ` 2.040-2.044 ns (1.96-3.91 GB/s) ` |
| ` 9-16 ` | ` 1.593-1.595 ns (5.64-10.04 GB/s) ` | ` 1.735-1.740 ns (5.19-9.20 GB/s) ` | ` 2.227-2.264 ns (4.04-7.07 GB/s) ` | ` 2.427-2.448 ns (3.71-6.53 GB/s) ` |
| ` 17-32 ` | ` 1.888-1.906 ns (8.92-16.95 GB/s) ` | ` 2.070-2.074 ns (8.21-15.43 GB/s) ` | ` 2.611-2.672 ns (6.36-12.26 GB/s) ` | ` 2.907-2.949 ns (5.76-11.01 GB/s) ` |
| ` 33-64 ` | ` 2.634-2.670 ns (12.36-24.30 GB/s) ` | ` 2.926-2.952 ns (11.18-21.88 GB/s) ` | ` 3.595-3.604 ns (9.18-17.76 GB/s) ` | ` 3.906-3.910 ns (8.45-16.37 GB/s) ` |
| ` 65-96 ` | ` 3.456-3.474 ns (18.81-27.63 GB/s) ` | ` 3.798-3.866 ns (17.12-24.84 GB/s) ` | ` 4.544-4.614 ns (14.30-20.81 GB/s) ` | ` 5.052-5.122 ns (12.86-18.74 GB/s) ` |
| ` 97-128 ` | ` 4.226-4.236 ns (22.96-30.21 GB/s) ` | ` 4.757-4.764 ns (20.36-26.91 GB/s) ` | ` 5.366-5.442 ns (18.08-23.52 GB/s) ` | ` 5.974-6.004 ns (16.15-21.43 GB/s) ` |

Ranges span measured endpoints (plus 2 B in the 1-3 B row), not every intervening size.

### Large Sizes

These are production-dispatch results, not refreshed forced-ISA columns.
A nonzero seed derives the 192-byte secret on each one-shot call above
240 B; Hasher retains its derived secret between writes.

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 129 ` | ` 4.984 ns (25.89 GB/s) ` | ` 5.506 ns (23.43 GB/s) ` | ` 6.840 ns (18.86 GB/s) ` | ` 7.662 ns (16.84 GB/s) ` |
| ` 240 ` | ` 7.858 ns (30.54 GB/s) ` | ` 8.570 ns (28.01 GB/s) ` | ` 9.790 ns (24.51 GB/s) ` | ` 10.765 ns (22.29 GB/s) ` |
| ` 241 ` | ` 8.644 ns (27.88 GB/s) ` | ` 13.930 ns (17.30 GB/s) ` | ` 10.010 ns (24.08 GB/s) ` | ` 15.090 ns (15.97 GB/s) ` |
| ` 512 ` | ` 12.845 ns (39.86 GB/s) ` | ` 16.875 ns (30.34 GB/s) ` | ` 14.010 ns (36.55 GB/s) ` | ` 18.410 ns (27.81 GB/s) ` |
| ` 1024 ` | ` 21.640 ns (47.32 GB/s) ` | ` 25.655 ns (39.91 GB/s) ` | ` 23.125 ns (44.28 GB/s) ` | ` 27.235 ns (37.60 GB/s) ` |
| ` 8192 ` | ` 157.250 ns (52.10 GB/s) ` | ` 163.750 ns (50.03 GB/s) ` | ` 159.200 ns (51.46 GB/s) ` | ` 164.950 ns (49.66 GB/s) ` |
| ` 102400 ` | ` 1.982 us (51.66 GB/s) ` | ` 1.971 us (51.95 GB/s) ` | ` 1.944 us (52.67 GB/s) ` | ` 1.956 us (52.37 GB/s) ` |
| ` 1024000 ` | ` 19.906 us (51.44 GB/s) ` | ` 19.880 us (51.51 GB/s) ` | ` 19.788 us (51.75 GB/s) ` | ` 19.852 us (51.58 GB/s) ` |
| ` 10240000 ` | ` 197.704 us (51.79 GB/s) ` | ` 196.958 us (51.99 GB/s) ` | ` 197.904 us (51.74 GB/s) ` | ` 198.412 us (51.61 GB/s) ` |
| ` 102400000 ` | ` 1.995 ms (51.32 GB/s) ` | ` 2.008 ms (50.99 GB/s) ` | ` 1.995 ms (51.32 GB/s) ` | ` 2.014 ms (50.85 GB/s) ` |

### Streaming

Each operation resets a reused Hasher, writes the complete input, and
computes Sum64 or Sum128. Seed setup is outside the timed loop.

| Bytes | Hasher64 | Hasher64 Seed | Hasher128 | Hasher128 Seed |
|---|---|---|---|---|
| ` 16 ` | ` 4.595 ns (3.48 GB/s) ` | ` 5.406 ns (2.96 GB/s) ` | ` 5.130 ns (3.12 GB/s) ` | ` 5.941 ns (2.69 GB/s) ` |
| ` 64 ` | ` 5.939 ns (10.78 GB/s) ` | ` 6.244 ns (10.25 GB/s) ` | ` 6.758 ns (9.47 GB/s) ` | ` 7.096 ns (9.02 GB/s) ` |
| ` 256 ` | ` 14.395 ns (17.78 GB/s) ` | ` 14.610 ns (17.52 GB/s) ` | ` 15.670 ns (16.34 GB/s) ` | ` 15.700 ns (16.31 GB/s) ` |
| ` 1024 ` | ` 34.925 ns (29.32 GB/s) ` | ` 34.895 ns (29.35 GB/s) ` | ` 36.375 ns (28.15 GB/s) ` | ` 36.205 ns (28.28 GB/s) ` |
| ` 4096 ` | ` 99.100 ns (41.33 GB/s) ` | ` 99.415 ns (41.20 GB/s) ` | ` 100.350 ns (40.82 GB/s) ` | ` 99.770 ns (41.05 GB/s) ` |
| ` 65536 ` | ` 1.282 us (51.12 GB/s) ` | ` 1.278 us (51.26 GB/s) ` | ` 1.270 us (51.60 GB/s) ` | ` 1.268 us (51.66 GB/s) ` |
| ` 1048576 ` | ` 20.291 us (51.68 GB/s) ` | ` 20.214 us (51.88 GB/s) ` | ` 20.275 us (51.72 GB/s) ` | ` 20.312 us (51.62 GB/s) ` |
| ` 16777216 ` | ` 332.993 us (50.38 GB/s) ` | ` 334.990 us (50.08 GB/s) ` | ` 329.864 us (50.86 GB/s) ` | ` 326.159 us (51.44 GB/s) ` |

### Cumulative Comparison

Time geomeans against cumulative baseline `0a183a1` are descriptive;
there is no aggregate p-value. Faster/slower counts use p < 0.05 per row.

| Class | Rows | Time change | Faster / slower |
|---|---:|---:|---:|
| One-shot 0-16 B | 32 | -2.76% | 14 / 7 |
| One-shot 17-128 B | 32 | -0.83% | 10 / 0 |
| One-shot 129-240 B | 8 | -5.62% | 7 / 0 |
| One-shot 241-1024 B | 12 | -7.76% | 12 / 0 |
| One-shot >1024 B | 20 | -0.78% | 3 / 0 |
| Hasher | 52 | -2.02% | 16 / 0 |
| Chunked Hasher | 6 | -0.06% | 0 / 0 |

The measured candidate is final source tip `71552bd`.
