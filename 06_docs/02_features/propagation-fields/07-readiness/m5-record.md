# M5's record: the globe frame on the reference machine

Written by `scripts/gate --cost` (plan task P7.1, watchpost D-350) on the reference machine, an Apple M5 Pro. A frame is
one cell's pan over the whole world under a 2° field filled and drawn with labelled contours, the night and
its terminator, and a reach's edge (`globePan`, `globe_bench_test.go`); each figure is the slower of the
field over the sea and stopping at the shore, each the median of 5 runs. 800 × 220 is recorded, a known
limit, not M5's (watchpost D-261). `scripts/gate --release` refuses a final v0.3 tag while this record is
missing, names other code or another machine, or misses a target.

| Field | Value |
|---|---|
| Code tree | `e5cb59079964` |
| Machine | Apple M5 Pro |
| Go | go1.26.10 darwin/arm64 |
| Measured | 2026-10-30T22:46:38Z |
| 400 × 110, ms a frame | 8.016 (target 16) |
| 200 × 56, ms a frame | 2.147 (target 8) |
| 800 × 220, ms a frame | 29.429 |
| Allocations a frame, 400 × 110 | 4201 |
