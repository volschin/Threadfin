//go:build linux && amd64

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type metricSummary struct {
	MedianPercentChange float64   `json:"median_percent_change"`
	MADPercent          float64   `json:"mad_percent"`
	ImprovedPairs       int       `json:"improved_pairs"`
	PairedChanges       []float64 `json:"paired_changes"`
}

type sessionSummary struct {
	Metrics          map[string]metricSummary `json:"metrics"`
	FailedPredicates []string                 `json:"failed_predicates"`
}

// percentageComparisonTolerance absorbs only floating-point representation
// error around exact engineering screen ratios, in percentage points.
const percentageComparisonTolerance = 1e-9

func atMostPercent(got, limit float64) bool {
	return got <= limit || math.Abs(got-limit) <= percentageComparisonTolerance
}

func atLeastPercent(got, limit float64) bool {
	return got >= limit || math.Abs(got-limit) <= percentageComparisonTolerance
}

func isLowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func parseUTCTimestamp(value string) (time.Time, bool) {
	if !strings.HasSuffix(value, "Z") {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil && !parsed.IsZero()
}

func isGo127Version(value string) bool {
	patch, found := strings.CutPrefix(value, "go1.27.")
	if !found || patch == "" || (len(patch) > 1 && patch[0] == '0') {
		return false
	}
	_, err := strconv.ParseUint(patch, 10, 32)
	return err == nil
}

func validResultContract(result runResult) bool {
	_, startedOK := parseUTCTimestamp(result.StartedAtUTC)
	_, guideStartOK := parseUTCTimestamp(result.GuideStartUTC)
	return result.SchemaVersion == 1 && startedOK && guideStartOK &&
		isLowerHex(result.GitCommit, 40) && isGo127Version(result.GoVersion) && result.GOAMD64 == "v1" &&
		isLowerHex(result.BinarySHA256, 64) && isLowerHex(result.PlaylistSHA256, 64) && isLowerHex(result.GuideSHA256, 64) &&
		(result.ProfileSHA256 == "" || isLowerHex(result.ProfileSHA256, 64))
}

func hasMinimumStreamSuccesses(attempts, successes int) bool {
	if attempts <= 0 || successes < 0 || successes > attempts {
		return false
	}
	// ceil(99% of attempts) is attempts - floor(attempts/100), which
	// avoids multiplying potentially large run counters.
	minimumSuccesses := attempts - attempts/100
	return successes >= minimumSuccesses
}

func summarizePairs(off, pgo [5]float64, higherIsBetter bool) (metricSummary, error) {
	changes := make([]float64, 5)
	improved := 0
	for i := range off {
		if off[i] <= 0 || pgo[i] <= 0 || math.IsNaN(off[i]) || math.IsNaN(pgo[i]) || math.IsInf(off[i], 0) || math.IsInf(pgo[i], 0) {
			return metricSummary{}, errors.New("metric values must be positive and finite")
		}
		changes[i] = 100 * (pgo[i]/off[i] - 1)
		if math.IsNaN(changes[i]) || math.IsInf(changes[i], 0) {
			return metricSummary{}, errors.New("metric percent changes must be finite")
		}
		if (higherIsBetter && changes[i] > 0) || (!higherIsBetter && changes[i] < 0) {
			improved++
		}
	}
	sorted := append([]float64(nil), changes...)
	slices.Sort(sorted)
	median := sorted[2]
	deviations := make([]float64, 5)
	for i, change := range changes {
		deviations[i] = math.Abs(change - median)
	}
	slices.Sort(deviations)
	return metricSummary{
		MedianPercentChange: median,
		MADPercent:          deviations[2],
		ImprovedPairs:       improved,
		PairedChanges:       changes,
	}, nil
}

// screenSession applies engineering screen thresholds, not inferred confidence bounds.
func screenSession(results []runResult) sessionSummary {
	summary := sessionSummary{Metrics: make(map[string]metricSummary)}
	fail := func(name string) { summary.FailedPredicates = append(summary.FailedPredicates, name) }
	if len(results) != 10 {
		fail("exactly_five_complete_pairs")
		return summary
	}

	var offByPair, pgoByPair [5]runResult
	var offSeen, pgoSeen [5]bool
	sequences := make(map[int]bool)
	commits, goVersions, goamd64s := map[string]bool{}, map[string]bool{}, map[string]bool{}
	playlists, guides, starts := map[string]bool{}, map[string]bool{}, map[string]bool{}
	offHashes, pgoHashes := map[string]bool{}, map[string]bool{}
	validRuns, validContracts := true, true
	for _, result := range results {
		if !validResultContract(result) {
			validContracts = false
			validRuns = false
		}
		if result.Pair < 1 || result.Pair > 5 || result.Sequence < 1 || result.Sequence > 10 || sequences[result.Sequence] {
			validRuns = false
			continue
		}
		sequences[result.Sequence] = true
		index := result.Pair - 1
		switch result.Variant {
		case "off":
			if offSeen[index] {
				validRuns = false
			}
			offSeen[index] = true
			offByPair[index] = result
			offHashes[result.BinarySHA256] = true
		case "pgo":
			if pgoSeen[index] {
				validRuns = false
			}
			pgoSeen[index] = true
			pgoByPair[index] = result
			pgoHashes[result.BinarySHA256] = true
		default:
			validRuns = false
		}
		commits[result.GitCommit], goVersions[result.GoVersion], goamd64s[result.GOAMD64] = true, true, true
		playlists[result.PlaylistSHA256], guides[result.GuideSHA256], starts[result.GuideStartUTC] = true, true, true
		if result.PlaylistEntries != playlistEntryCount || result.XMLTVChannels != xmltvChannelCount || result.XMLTVPrograms != xmltvProgramCount ||
			!hasMinimumStreamSuccesses(result.StreamAttempts, result.StreamSuccesses) || result.StreamBytes < 256<<20 || result.ProfileStopReason != "signal" {
			validRuns = false
		}
	}
	for i := range offSeen {
		validRuns = validRuns && offSeen[i] && pgoSeen[i]
		if !offSeen[i] || !pgoSeen[i] {
			continue
		}
		first, second := offByPair[i].Sequence, pgoByPair[i].Sequence
		if i%2 == 0 {
			validRuns = validRuns && first == 2*i+1 && second == 2*i+2
		} else {
			validRuns = validRuns && second == 2*i+1 && first == 2*i+2
		}
	}
	if !validContracts {
		fail("valid_result_contract")
	}
	if !validRuns {
		fail("valid_complete_runs")
	}
	if len(commits) != 1 || len(goVersions) != 1 || len(goamd64s) != 1 || !goamd64s["v1"] {
		fail("identical_source_toolchain")
	}
	if len(playlists) != 1 || len(guides) != 1 || len(starts) != 1 {
		fail("identical_fixtures")
	}
	if len(offHashes) != 1 || len(pgoHashes) != 1 {
		fail("reproducible_variant_binaries")
	}
	if len(summary.FailedPredicates) != 0 {
		slices.Sort(summary.FailedPredicates)
		return summary
	}

	metric := func(name string, higher bool, get func(runResult) float64) metricSummary {
		var off, pgo [5]float64
		for i := range off {
			off[i], pgo[i] = get(offByPair[i]), get(pgoByPair[i])
		}
		got, err := summarizePairs(off, pgo, higher)
		if err != nil {
			fail("positive_finite_" + name)
		}
		summary.Metrics[name] = got
		return got
	}
	playlist := metric("playlist_throughput", true, func(r runResult) float64 { return r.PlaylistEntriesPerSecond })
	xepg := metric("xepg_throughput", true, func(r runResult) float64 { return r.XEPGProgramsPerSecond })
	stream := metric("stream_throughput", true, func(r runResult) float64 { return r.StreamBytesPerSecond })
	cpu := metric("cpu", false, func(r runResult) float64 { return r.ThreadfinCPUSeconds })
	ttfb := metric("ttfb_p95", false, func(r runResult) float64 { return r.StreamTTFBP95Milliseconds })
	rss := metric("peak_rss", false, func(r runResult) float64 { return float64(r.PeakRSSBytes) })
	alloc := metric("total_alloc", false, func(r runResult) float64 { return float64(r.TotalAllocBytes) })
	mallocs := metric("mallocs", false, func(r runResult) float64 { return float64(r.Mallocs) })
	pause := metric("gc_pause", false, func(r runResult) float64 { return float64(r.PauseTotalNanoseconds) })
	gc := metric("gc_cycles", false, func(r runResult) float64 { return float64(r.NumGC) })
	size := metric("binary_size", false, func(r runResult) float64 { return float64(r.BinarySizeBytes) })

	if !atMostPercent(cpu.MedianPercentChange, -3) || cpu.ImprovedPairs < 4 {
		fail("cpu_improves_3_percent_in_4_pairs")
	}
	qualifying := 0
	for _, value := range []metricSummary{playlist, xepg, stream} {
		if atLeastPercent(value.MedianPercentChange, 3) && value.ImprovedPairs >= 4 {
			qualifying++
		}
		if !atLeastPercent(value.MedianPercentChange, -1) {
			fail("no_throughput_regression_over_1_percent")
		}
	}
	if qualifying < 2 {
		fail("two_throughputs_improve_3_percent_in_4_pairs")
	}
	for name, value := range map[string]metricSummary{"ttfb": ttfb, "rss": rss, "alloc": alloc, "mallocs": mallocs, "pause": pause} {
		if !atMostPercent(value.MedianPercentChange, 2) {
			fail(name + "_regression_at_most_2_percent")
		}
	}
	if !atMostPercent(gc.MedianPercentChange, 5) {
		fail("gc_cycle_regression_at_most_5_percent")
	}
	if !atMostPercent(size.MedianPercentChange, 5) {
		fail("binary_growth_at_most_5_percent")
	}
	slices.Sort(summary.FailedPredicates)
	return summary
}

type sessionProvenance struct {
	GitCommit      string
	GoVersion      string
	GOAMD64        string
	OffBinary      string
	PGOBinary      string
	PlaylistSHA256 string
	GuideSHA256    string
	GuideStartUTC  string
	FirstStartedAt time.Time
	LastStartedAt  time.Time
}

func provenanceForSession(results []runResult) (sessionProvenance, bool) {
	if len(results) != 10 {
		return sessionProvenance{}, false
	}
	ordered := append([]runResult(nil), results...)
	slices.SortFunc(ordered, func(a, b runResult) int { return a.Sequence - b.Sequence })
	provenance := sessionProvenance{
		GitCommit:      ordered[0].GitCommit,
		GoVersion:      ordered[0].GoVersion,
		GOAMD64:        ordered[0].GOAMD64,
		PlaylistSHA256: ordered[0].PlaylistSHA256,
		GuideSHA256:    ordered[0].GuideSHA256,
		GuideStartUTC:  ordered[0].GuideStartUTC,
	}
	for index, result := range ordered {
		if result.Sequence != index+1 || result.GitCommit != provenance.GitCommit || result.GoVersion != provenance.GoVersion ||
			result.GOAMD64 != provenance.GOAMD64 || result.PlaylistSHA256 != provenance.PlaylistSHA256 ||
			result.GuideSHA256 != provenance.GuideSHA256 || result.GuideStartUTC != provenance.GuideStartUTC {
			return sessionProvenance{}, false
		}
		startedAt, timestampOK := parseUTCTimestamp(result.StartedAtUTC)
		if !timestampOK || (index > 0 && !startedAt.After(provenance.LastStartedAt)) {
			return sessionProvenance{}, false
		}
		if index == 0 {
			provenance.FirstStartedAt = startedAt
		}
		provenance.LastStartedAt = startedAt
		switch result.Variant {
		case "off":
			if provenance.OffBinary != "" && provenance.OffBinary != result.BinarySHA256 {
				return sessionProvenance{}, false
			}
			provenance.OffBinary = result.BinarySHA256
		case "pgo":
			if provenance.PGOBinary != "" && provenance.PGOBinary != result.BinarySHA256 {
				return sessionProvenance{}, false
			}
			provenance.PGOBinary = result.BinarySHA256
		default:
			return sessionProvenance{}, false
		}
	}
	return provenance, provenance.OffBinary != "" && provenance.PGOBinary != ""
}

func matchingOrderedSessions(first, second []runResult) bool {
	firstProvenance, firstOK := provenanceForSession(first)
	secondProvenance, secondOK := provenanceForSession(second)
	return firstOK && secondOK && secondProvenance.FirstStartedAt.After(firstProvenance.LastStartedAt) &&
		firstProvenance.GitCommit == secondProvenance.GitCommit &&
		firstProvenance.GoVersion == secondProvenance.GoVersion &&
		firstProvenance.GOAMD64 == secondProvenance.GOAMD64 &&
		firstProvenance.OffBinary == secondProvenance.OffBinary &&
		firstProvenance.PGOBinary == secondProvenance.PGOBinary &&
		firstProvenance.PlaylistSHA256 == secondProvenance.PlaylistSHA256 &&
		firstProvenance.GuideSHA256 == secondProvenance.GuideSHA256 &&
		firstProvenance.GuideStartUTC == secondProvenance.GuideStartUTC
}

type comparison struct {
	Verdict  string           `json:"verdict"`
	Sessions []sessionSummary `json:"sessions"`
}

func readSession(path string) ([]runResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	results := make([]runResult, 0, 10)
	for scanner.Scan() {
		var result runResult
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, scanner.Err()
}

func sameSessionFile(firstPath, secondPath string) bool {
	if filepath.Clean(firstPath) == filepath.Clean(secondPath) {
		return true
	}
	firstInfo, firstErr := os.Stat(firstPath)
	secondInfo, secondErr := os.Stat(secondPath)
	return firstErr == nil && secondErr == nil && os.SameFile(firstInfo, secondInfo)
}

func compareSessions(firstPath, secondPath, outputPath string) (comparison, error) {
	if !filepath.IsAbs(firstPath) || !filepath.IsAbs(secondPath) || !filepath.IsAbs(outputPath) {
		return comparison{}, errors.New("session and comparison paths must be absolute")
	}
	first, err := readSession(firstPath)
	if err != nil {
		return comparison{}, err
	}
	second, err := readSession(secondPath)
	if err != nil {
		return comparison{}, err
	}
	result := comparison{Verdict: "NO_ADOPT", Sessions: []sessionSummary{screenSession(first), screenSession(second)}}
	if len(result.Sessions[0].FailedPredicates) == 0 && len(result.Sessions[1].FailedPredicates) == 0 &&
		!sameSessionFile(firstPath, secondPath) && matchingOrderedSessions(first, second) {
		result.Verdict = "PILOT_SUPPORTS_SEPARATE_OPERATIONAL_REVIEW"
	}
	body, err := json.Marshal(result)
	if err != nil {
		return comparison{}, err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return comparison{}, err
	}
	tmp := outputPath + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return comparison{}, err
	}
	_, writeErr := file.Write(body)
	closeErr := errors.Join(file.Sync(), file.Close())
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(tmp)
		return comparison{}, err
	}
	if err := os.Rename(tmp, outputPath); err != nil {
		_ = os.Remove(tmp)
		return comparison{}, err
	}
	return result, nil
}
