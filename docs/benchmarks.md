# Threadfin Benchmarks

Threadfin's benchmarks provide evidence for performance-sensitive source changes.
They are not shared-runner timing gates.

## Requirements

- Go 1.27.0 or newer, matching `go.mod`.
- Vendored module mode.
- `benchstat` from `golang.org/x/perf/cmd/benchstat`.
- Before and after samples collected on the same machine under comparable load.

Install `benchstat` outside this module when it is not already available:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Compile Without Running

```bash
go test -mod=vendor -run='^$' -bench='^$' ./...
```

This is the CI contract. Compilation failures are gating; timing values are not.
Before gated Task 1B lands, the parser benchmark file defines only the normal
benchmark. After its separately approved correctness prerequisite and Task 1B
land, this same command also compiles `BenchmarkM3UParsingLarge` without
executing its body.

## Collect Before Samples

Run from the repository root on the benchmark-foundation commit, before one
performance-sensitive production change:

```bash
CORE_RE='^Benchmark(M3UParsingNormal|FilterConditions|SegmentTransfer|CachedImageDelivery|APIJSONDecoding|UpdateJSONDecoding|MapToJSON|Argon2HashSerialization)$'

go test -mod=vendor -run='^$' -bench="$CORE_RE" \
  -benchmem -count=10 -timeout=60m ./src/... > before.txt

go test -mod=vendor -run='^$' \
  -bench='^BenchmarkXMLTVGeneration$/^10Channels_100Programs$' \
  -benchmem -count=10 -timeout=60m ./src >> before.txt

go test -mod=vendor -run='^$' \
  -bench='^BenchmarkXMLTVGeneration$/^100Channels_10000Programs$' \
  -benchtime=1x -benchmem -count=10 -timeout=90m ./src >> before.txt
```

The last command deliberately takes ten one-iteration samples so the required
100-channel/10,000-program case does not inherit the default adaptive benchmark
duration.

## Collect After Samples

After applying exactly one performance-sensitive production change, on the same
machine and under comparable load:

```bash
CORE_RE='^Benchmark(M3UParsingNormal|FilterConditions|SegmentTransfer|CachedImageDelivery|APIJSONDecoding|UpdateJSONDecoding|MapToJSON|Argon2HashSerialization)$'

go test -mod=vendor -run='^$' -bench="$CORE_RE" \
  -benchmem -count=10 -timeout=60m ./src/... > after.txt

go test -mod=vendor -run='^$' \
  -bench='^BenchmarkXMLTVGeneration$/^10Channels_100Programs$' \
  -benchmem -count=10 -timeout=60m ./src >> after.txt

go test -mod=vendor -run='^$' \
  -bench='^BenchmarkXMLTVGeneration$/^100Channels_10000Programs$' \
  -benchtime=1x -benchmem -count=10 -timeout=90m ./src >> after.txt
```

## Add Gated Large-M3U Samples

Do not run these commands while
`src/internal/m3u-parser/xteve_m3uParser_optimized.go` still contains
`if len(lines) < 2 {`. The current implementation returns zero streams for the
ordinary one-URL fixtures, and the benchmark intentionally rejects that defect.

Only after the separately approved correctness fix and Task 1B land, append ten
large-parser samples. Run the `before.txt` command on the gated benchmark-foundation
commit. Check out or apply exactly one later performance change, rerun the gate,
and then run the `after.txt` command on that changed commit:

```bash
! git grep -n -F 'if len(lines) < 2 {' -- src/internal/m3u-parser/xteve_m3uParser_optimized.go

go test -mod=vendor -run='^$' -bench='^BenchmarkM3UParsingLarge$' \
  -benchmem -count=10 -timeout=60m ./src/internal/m3u-parser >> before.txt

! git grep -n -F 'if len(lines) < 2 {' -- src/internal/m3u-parser/xteve_m3uParser_optimized.go

go test -mod=vendor -run='^$' -bench='^BenchmarkM3UParsingLarge$' \
  -benchmem -count=10 -timeout=60m ./src/internal/m3u-parser >> after.txt
```

The leading gate command must exit successfully with no output. Every large
sub-benchmark must report its requested stream count; zero-output samples are
invalid and must not enter `benchstat`.

## Compare

```bash
benchstat before.txt after.txt
```

Review `ns/op`, `B/op`, and `allocs/op` together. Retain an optimization only
when the relevant benchmark improves without changing observable behavior.
Shared-runner wall-clock variation is informational and never a merge gate.
Deterministic allocation assertions require a separate change after repeated
observations establish stable counts.
