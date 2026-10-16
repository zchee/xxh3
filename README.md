# XXH3
[![GoDoc](https://godoc.org/github.com/zeebo/xxh3?status.svg)](https://godoc.org/github.com/zeebo/xxh3)
[![Sourcegraph](https://sourcegraph.com/github.com/zeebo/xxh3/-/badge.svg)](https://sourcegraph.com/github.com/zeebo/xxh3?badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/zeebo/xxh3)](https://goreportcard.com/report/github.com/zeebo/xxh3)

This package is a port of the [xxh3](https://github.com/Cyan4973/xxHash) library to Go.

Upstream has fixed the output as of v0.8.0, and this package matches that.

---

# Benchmarks

The AMD Ryzen 9 9950X and Intel i7-8850H measurements below predate this branch's changes; that hardware was not re-measured here.

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

### Small Sizes

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 0 ` | ` 3.159 ns ` | ` 3.046 ns ` | ` 2.707 ns ` | ` 3.687 ns ` |
| ` 1-3 ` | ` 2.705-3.095 ns (0.36-1.11 GB/s) ` | ` 3.014-3.032 ns (0.33-0.99 GB/s) ` | ` 3.355-3.511 ns (0.28-0.89 GB/s) ` | ` 4.021-4.158 ns (0.25-0.75 GB/s) ` |
| ` 4-8 ` | ` 2.682-2.684 ns (1.49-2.98 GB/s) ` | ` 3.128 ns (1.28-2.56 GB/s) ` | ` 3.349-3.350 ns (1.19-2.39 GB/s) ` | ` 3.806 ns (1.05-2.10 GB/s) ` |
| ` 9-16 ` | ` 2.688 ns (3.35-5.95 GB/s) ` | ` 2.777-2.784 ns (3.23-5.76 GB/s) ` | ` 3.942-3.943 ns (2.28-4.06 GB/s) ` | ` 4.304 ns (2.09-3.72 GB/s) ` |
| ` 17-32 ` | ` 3.950-3.978 ns (4.27-8.10 GB/s) ` | ` 3.379-3.382 ns (5.03-9.46 GB/s) ` | ` 4.944-4.945 ns (3.44-6.47 GB/s) ` | ` 5.192 ns (3.27-6.16 GB/s) ` |
| ` 33-64 ` | ` 4.955-4.979 ns (6.66-12.85 GB/s) ` | ` 4.747-4.763 ns (6.93-13.48 GB/s) ` | ` 6.899-6.930 ns (4.76-9.28 GB/s) ` | ` 7.171-7.258 ns (4.55-8.92 GB/s) ` |
| ` 65-96 ` | ` 5.873-5.875 ns (11.07-16.34 GB/s) ` | ` 6.321-6.323 ns (10.28-15.19 GB/s) ` | ` 8.726-8.727 ns (7.45-11.00 GB/s) ` | ` 9.238-9.240 ns (7.03-10.39 GB/s) ` |
| ` 97-128 ` | ` 7.380-7.386 ns (13.13-17.34 GB/s) ` | ` 7.716-7.742 ns (12.53-16.59 GB/s) ` | ` 10.650-10.660 ns (9.10-12.02 GB/s) ` | ` 11.130-11.160 ns (8.69-11.50 GB/s) ` |

Ranges span measured endpoints (plus 2 B in the 1-3 B row), not every intervening size.

Measured on Intel Xeon Platinum 8481C (linux/amd64, Sapphire Rapids VM), Go 1.27.2, 2026-10-16 20:25:48 UTC; 10 alternating samples, cpu=1, 300 ms per leaf.

### Large Sizes

64-bit unseeded hashing is shown. Scalar, SSE2, AVX2 and AVX512 columns
force their respective kernels; the scalar column is not the normal SIMD
dispatch. Blank cells have no separately measured SIMD row.

| Bytes | Scalar Rate | SSE2 Rate | AVX2 Rate | AVX512 Rate |
|---|---|---|---|---|
| ` 129 ` | ` 8.820 ns/op (14.63 GB/s) ` |  |  |  |
| ` 240 ` | ` 13.470 ns/op (17.82 GB/s) ` |  |  |  |
| ` 241 ` | ` 27.390 ns/op (8.80 GB/s) ` | ` 14.540 ns/op (16.57 GB/s) ` | ` 11.900 ns/op (20.25 GB/s) ` |  |
| ` 512 ` | ` 47.090 ns/op (10.87 GB/s) ` | ` 22.280 ns/op (22.98 GB/s) ` | ` 16.720 ns/op (30.62 GB/s) ` |  |
| ` 1024 ` | ` 86.700 ns/op (11.81 GB/s) ` | ` 43.050 ns/op (23.79 GB/s) ` | ` 26.610 ns/op (38.48 GB/s) ` | ` 24.070 ns/op (42.54 GB/s) ` |
| ` 100KB ` | ` 8.490 us/op (12.06 GB/s) ` | ` 4.063 us/op (25.20 GB/s) ` | ` 2.252 us/op (45.47 GB/s) ` | ` 1.685 us/op (60.77 GB/s) ` |

A non-zero seed re-derives the 192-byte secret on each one-shot call above
240 B. At 512 B, SSE2 takes 37.62 ns seeded versus 22.28 ns unseeded. At
100 KB (102400 B), AVX512 takes 1.714 us versus 1.685 us, a 1.7% difference.
Hasher retains its derived secret and does not have the same per-call cost.

Go/C time ratios (unseeded 64-bit, AVX512, fresh same-machine C reference): 1 KiB = 1.022; 100 KB (102400 B) = 0.829; lower is better.

Against the pre-optimization baseline, the 240-row time geomean improves
10.44%, but 20 rows are significantly slower (p < 0.05).
The 17 B and 32 B unseeded 64-bit losses (+21.65% and +21.07%), and the
33 B and 64 B losses, are benchmark-loop code-placement effects in the test
package, not hash-function changes: the hashing code is identical and
identically aligned to the previous version.
With that same hashing code and the previous test file, 17 B and 32 B each
measure 3.11 ns, and 33 B / 64 B measure 4.51 / 4.38 ns, matching the earlier
values.
The 0 B unseeded 128-bit row is +11.65%, with no identified cause;
benchstat reports ±9% uncertainty in the focused comparison (cumulative
raw samples 2.540-2.920 ns), and the remaining slower rows range from
+0.16% to +2.43%.

Measured on Intel Xeon Platinum 8481C (linux/amd64, Sapphire Rapids VM), Go 1.27.2, 2026-10-16 20:25:48 UTC; 10 alternating samples, cpu=1, 300 ms per leaf.

## Apple M3 Max (darwin/arm64)

### Small Sizes

| Bytes | 64-bit | 64-bit Seed | 128-bit | 128-bit Seed |
|---|---|---|---|---|
| ` 0 ` | ` 2.464 ns ` | ` 2.722 ns ` | ` 2.447 ns ` | ` 3.008 ns ` |
| ` 1-3 ` | ` 2.177-2.436 ns (0.41-1.38 GB/s) ` | ` 2.174-2.474 ns (0.41-1.38 GB/s) ` | ` 2.171-2.707 ns (0.37-1.38 GB/s) ` | ` 2.456-2.701 ns (0.37-1.22 GB/s) ` |
| ` 4-8 ` | ` 2.163-2.175 ns (1.84-3.70 GB/s) ` | ` 2.158-2.177 ns (1.84-3.71 GB/s) ` | ` 2.471-2.494 ns (1.60-3.24 GB/s) ` | ` 2.444-2.452 ns (1.64-3.26 GB/s) ` |
| ` 9-16 ` | ` 2.177-2.180 ns (4.13-7.34 GB/s) ` | ` 2.440-2.493 ns (3.61-6.56 GB/s) ` | ` 2.717-2.732 ns (3.29-5.89 GB/s) ` | ` 2.736-2.765 ns (3.25-5.85 GB/s) ` |
| ` 17-32 ` | ` 2.250-2.252 ns (7.56-14.21 GB/s) ` | ` 2.431-2.470 ns (6.88-13.16 GB/s) ` | ` 2.954-2.956 ns (5.75-10.83 GB/s) ` | ` 3.206-3.218 ns (5.28-9.98 GB/s) ` |
| ` 33-64 ` | ` 2.972-2.993 ns (11.10-21.38 GB/s) ` | ` 3.265-3.270 ns (10.11-19.57 GB/s) ` | ` 3.894-3.943 ns (8.37-16.44 GB/s) ` | ` 4.304-4.365 ns (7.56-14.87 GB/s) ` |
| ` 65-96 ` | ` 3.788-3.816 ns (17.16-25.16 GB/s) ` | ` 4.212-4.217 ns (15.41-22.79 GB/s) ` | ` 4.938-4.942 ns (13.16-19.43 GB/s) ` | ` 5.386-5.423 ns (12.07-17.70 GB/s) ` |
| ` 97-128 ` | ` 4.496-4.592 ns (21.12-28.47 GB/s) ` | ` 5.018-5.079 ns (19.10-25.51 GB/s) ` | ` 5.767-5.788 ns (16.76-22.20 GB/s) ` | ` 6.358-6.390 ns (15.26-20.03 GB/s) ` |

Ranges span measured endpoints (plus 2 B in the 1-3 B row), not every intervening size.

Measured on Apple M3 Max (darwin/arm64), Go 1.27.2, 2026-10-17 05:26:29 JST; 10 alternating samples, cpu=1, 300 ms per leaf.

### Large Sizes

64-bit unseeded default is shown; inputs above 240 B use NEON.

| Bytes | NEON (default) Rate |
|---|---|
| ` 129 ` | ` 5.352 ns/op (24.10 GB/s) ` |
| ` 240 ` | ` 9.104 ns/op (26.36 GB/s) ` |
| ` 241 ` | ` 9.815 ns/op (24.55 GB/s) ` |
| ` 512 ` | ` 13.950 ns/op (36.70 GB/s) ` |
| ` 1024 ` | ` 23.230 ns/op (44.08 GB/s) ` |
| ` 100KB ` | ` 1.956 us/op (52.35 GB/s) ` |

Go/C time ratios (unseeded 64-bit, NEON/default, fresh same-machine C reference): 1 KiB = 0.891; 100 KB (102400 B) = 0.803; lower is better.

Against the pre-optimization baseline, the 116-row time geomean improves
21.89%, with no significantly slower row (p < 0.05).

Measured on Apple M3 Max (darwin/arm64), Go 1.27.2, 2026-10-17 05:26:29 JST; 10 alternating samples, cpu=1, 300 ms per leaf.
