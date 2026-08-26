//go:build linux && amd64

package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func baselineResult(variant string, pair, sequence int) runResult {
	result := runResult{
		SchemaVersion: 1, Variant: variant, Pair: pair, Sequence: sequence,
		GitCommit: strings.Repeat("a", 40), GoVersion: "go1.27.0", GOAMD64: "v1",
		BinarySHA256:    strings.Repeat(map[string]string{"off": "b", "pgo": "c"}[variant], 64),
		BinarySizeBytes: 10_000_000,
		PlaylistSHA256:  strings.Repeat("d", 64), GuideSHA256: strings.Repeat("e", 64),
		GuideStartUTC:   "2026-08-27T00:00:00Z",
		PlaylistEntries: 100_000, XMLTVChannels: 100, XMLTVPrograms: 10_000,
		PlaylistEntriesPerSecond: 10_000, XEPGProgramsPerSecond: 2_000,
		StreamAttempts: 100, StreamSuccesses: 100, StreamBytes: 300 << 20,
		StreamBytesPerSecond: 8 << 20, StreamTTFBP95Milliseconds: 100,
		ThreadfinCPUSeconds: 100, PeakRSSBytes: 100 << 20,
		TotalAllocBytes: 1_000_000_000, Mallocs: 1_000_000, Frees: 900_000,
		NumGC: 100, PauseTotalNanoseconds: 10_000_000, HeapAllocBytesAtStop: 50 << 20,
		ProfileStopReason: "signal",
	}
	if variant == "pgo" {
		result.PlaylistEntriesPerSecond *= 1.04
		result.XEPGProgramsPerSecond *= 1.04
		result.StreamBytesPerSecond *= 1.04
		result.ThreadfinCPUSeconds *= 0.96
	}
	return result
}

func passingSession() []runResult {
	results := make([]runResult, 0, 10)
	sequence := 0
	for pair := 1; pair <= 5; pair++ {
		order := []string{"off", "pgo"}
		if pair%2 == 0 {
			order = []string{"pgo", "off"}
		}
		for _, variant := range order {
			sequence++
			results = append(results, baselineResult(variant, pair, sequence))
		}
	}
	return results
}

func TestCompareScreen(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func([]runResult)
		wantFailure string
	}{
		{name: "pass", mutate: func([]runResult) {}, wantFailure: ""},
		{name: "cpu threshold", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].ThreadfinCPUSeconds = 98
				}
			}
		}, wantFailure: "cpu_improves_3_percent_in_4_pairs"},
		{name: "stream success", mutate: func(rows []runResult) { rows[0].StreamSuccesses = 98 }, wantFailure: "valid_complete_runs"},
		{name: "stream success overflow", mutate: func(rows []runResult) {
			rows[0].StreamAttempts = int(^uint(0) >> 1)
			rows[0].StreamSuccesses = 0
		}, wantFailure: "valid_complete_runs"},
		{name: "binary growth", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].BinarySizeBytes = 10_500_001
				}
			}
		}, wantFailure: "binary_growth_at_most_5_percent"},
		{name: "fixture mismatch", mutate: func(rows []runResult) { rows[0].GuideSHA256 = strings.Repeat("f", 64) }, wantFailure: "identical_fixtures"},
		{name: "missing pair", mutate: func(rows []runResult) { rows[0].Pair = 0 }, wantFailure: "valid_complete_runs"},
		{name: "missing sequence", mutate: func(rows []runResult) { rows[0].Sequence = 0 }, wantFailure: "valid_complete_runs"},
		{name: "duplicate variant", mutate: func(rows []runResult) { rows[0].Variant = "pgo" }, wantFailure: "valid_complete_runs"},
		{name: "timeout", mutate: func(rows []runResult) { rows[0].ProfileStopReason = "timeout" }, wantFailure: "valid_complete_runs"},
		{name: "cardinality", mutate: func(rows []runResult) { rows[0].XMLTVPrograms = 9999 }, wantFailure: "valid_complete_runs"},
		{name: "minimum bytes", mutate: func(rows []runResult) { rows[0].StreamBytes = (256 << 20) - 1 }, wantFailure: "valid_complete_runs"},
		{name: "toolchain", mutate: func(rows []runResult) { rows[0].GoVersion = "go1.27.1" }, wantFailure: "identical_source_toolchain"},
		{name: "unstable hash", mutate: func(rows []runResult) { rows[0].BinarySHA256 = strings.Repeat("9", 64) }, wantFailure: "reproducible_variant_binaries"},
		{name: "throughput regression", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].StreamBytesPerSecond = float64(8<<20) * .98
				}
			}
		}, wantFailure: "no_throughput_regression_over_1_percent"},
		{name: "ttfb", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].StreamTTFBP95Milliseconds *= 1.021
				}
			}
		}, wantFailure: "ttfb_regression_at_most_2_percent"},
		{name: "rss", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PeakRSSBytes = 103 << 20
				}
			}
		}, wantFailure: "rss_regression_at_most_2_percent"},
		{name: "alloc", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].TotalAllocBytes = 1_021_000_000
				}
			}
		}, wantFailure: "alloc_regression_at_most_2_percent"},
		{name: "mallocs", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].Mallocs = 1_021_000
				}
			}
		}, wantFailure: "mallocs_regression_at_most_2_percent"},
		{name: "gc pause", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PauseTotalNanoseconds = 10_210_000
				}
			}
		}, wantFailure: "pause_regression_at_most_2_percent"},
		{name: "gc cycles", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].NumGC = 106
				}
			}
		}, wantFailure: "gc_cycle_regression_at_most_5_percent"},
		{name: "cpu pair direction", mutate: func(rows []runResult) {
			changed := 0
			for i := range rows {
				if rows[i].Variant == "pgo" && changed < 2 {
					rows[i].ThreadfinCPUSeconds = 101
					changed++
				}
			}
		}, wantFailure: "cpu_improves_3_percent_in_4_pairs"},
		{name: "throughput qualification", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PlaylistEntriesPerSecond = 10_200
					rows[i].XEPGProgramsPerSecond = 2_040
					rows[i].StreamBytesPerSecond = 8 << 20
				}
			}
		}, wantFailure: "two_throughputs_improve_3_percent_in_4_pairs"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := passingSession()
			test.mutate(rows)
			got := screenSession(rows)
			if test.wantFailure == "" {
				if len(got.FailedPredicates) != 0 {
					t.Fatalf("failures = %v", got.FailedPredicates)
				}
				return
			}
			if !slices.Contains(got.FailedPredicates, test.wantFailure) {
				t.Fatalf("failures = %v, want %s", got.FailedPredicates, test.wantFailure)
			}
		})
	}
}

func TestCompareScreenAllowsExactBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]runResult)
	}{
		{name: "99 percent stream success", mutate: func(rows []runResult) {
			rows[0].StreamAttempts, rows[0].StreamSuccesses = 100, 99
		}},
		{name: "256 MiB stream bytes", mutate: func(rows []runResult) { rows[0].StreamBytes = 256 << 20 }},
		{name: "cpu minus 3 percent in four pairs", mutate: func(rows []runResult) {
			unchanged := true
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].ThreadfinCPUSeconds = 97
					if unchanged {
						rows[i].ThreadfinCPUSeconds = 100
						unchanged = false
					}
				}
			}
		}},
		{name: "two throughputs plus 3 percent in four pairs", mutate: func(rows []runResult) {
			unchanged := true
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PlaylistEntriesPerSecond = 10_300
					rows[i].XEPGProgramsPerSecond = 2_060
					rows[i].StreamBytesPerSecond = 8 << 20
					if unchanged {
						rows[i].PlaylistEntriesPerSecond = 10_000
						rows[i].XEPGProgramsPerSecond = 2_000
						unchanged = false
					}
				}
			}
		}},
		{name: "throughput minus 1 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].StreamBytesPerSecond = float64(8<<20) * .99
				}
			}
		}},
		{name: "ttfb plus 2 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].StreamTTFBP95Milliseconds = 102
				}
			}
		}},
		{name: "rss plus 2 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PeakRSSBytes = 102 << 20
				}
			}
		}},
		{name: "allocation plus 2 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].TotalAllocBytes = 1_020_000_000
				}
			}
		}},
		{name: "mallocs plus 2 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].Mallocs = 1_020_000
				}
			}
		}},
		{name: "gc pause plus 2 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].PauseTotalNanoseconds = 10_200_000
				}
			}
		}},
		{name: "gc cycles plus 5 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].NumGC = 105
				}
			}
		}},
		{name: "binary growth plus 5 percent", mutate: func(rows []runResult) {
			for i := range rows {
				if rows[i].Variant == "pgo" {
					rows[i].BinarySizeBytes = 10_500_000
				}
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rows := passingSession()
			test.mutate(rows)
			if got := screenSession(rows); len(got.FailedPredicates) != 0 {
				t.Fatalf("failures = %v", got.FailedPredicates)
			}
		})
	}
}

func TestCompareScreenRejectsNonTenRowSessions(t *testing.T) {
	tests := []struct {
		name string
		rows func() []runResult
	}{
		{name: "nine rows", rows: func() []runResult { return passingSession()[:9] }},
		{name: "eleven rows", rows: func() []runResult {
			rows := passingSession()
			return append(rows, baselineResult("off", 1, 1))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := screenSession(test.rows())
			if !slices.Contains(got.FailedPredicates, "exactly_five_complete_pairs") || len(got.Metrics) != 0 {
				t.Fatalf("summary = %#v", got)
			}
		})
	}
}

func TestCompareInputOrderDoesNotChangeSummary(t *testing.T) {
	ordered := screenSession(passingSession())
	shuffled := passingSession()
	slices.Reverse(shuffled)
	got := screenSession(shuffled)
	orderedJSON, err := json.Marshal(ordered)
	if err != nil {
		t.Fatal(err)
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotJSON) != string(orderedJSON) {
		t.Fatalf("shuffled summary = %s, want %s", gotJSON, orderedJSON)
	}
}

func TestCompareSessionsRequiresBothPassingSessions(t *testing.T) {
	directory := t.TempDir()
	firstPath := filepath.Join(directory, "first.jsonl")
	secondPath := filepath.Join(directory, "second.jsonl")
	outputPath := filepath.Join(directory, "comparison.json")
	writeSession(t, firstPath, passingSession())
	writeSession(t, secondPath, passingSession())

	got, err := compareSessions(firstPath, secondPath, outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "PILOT_SUPPORTS_SEPARATE_OPERATIONAL_REVIEW" {
		t.Fatalf("verdict = %q", got.Verdict)
	}

	failing := passingSession()
	failing[0].ProfileStopReason = "timeout"
	writeSession(t, secondPath, failing)
	got, err = compareSessions(firstPath, secondPath, outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.Verdict != "NO_ADOPT" {
		t.Fatalf("verdict = %q", got.Verdict)
	}
	if len(got.Sessions[0].FailedPredicates) != 0 || len(got.Sessions[1].FailedPredicates) == 0 {
		t.Fatalf("session failures = %#v", got.Sessions)
	}
}

func TestSummarizePairsRejectsNonPositiveAndNonFiniteValues(t *testing.T) {
	valid := [5]float64{1, 1, 1, 1, 1}
	for _, invalid := range []float64{0, -1, math.Inf(1), math.NaN()} {
		off := valid
		off[0] = invalid
		if _, err := summarizePairs(off, valid, true); err == nil {
			t.Fatalf("invalid denominator %v was accepted", invalid)
		}
		pgo := valid
		pgo[0] = invalid
		if _, err := summarizePairs(valid, pgo, true); err == nil {
			t.Fatalf("invalid numerator %v was accepted", invalid)
		}
	}
}

func writeSession(t *testing.T, path string, rows []runResult) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		body, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(append(body, '\n')); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
