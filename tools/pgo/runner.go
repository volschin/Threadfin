//go:build linux && amd64

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type runConfig struct {
	Variant        string
	Binary         string
	Output         string
	Profile        string
	Pair           int
	Sequence       int
	StreamDuration time.Duration
	Clients        int
	SampleBytes    int64
	GuideStart     time.Time
}

type runResult struct {
	SchemaVersion             int     `json:"schema_version"`
	Variant                   string  `json:"variant"`
	Pair                      int     `json:"pair"`
	Sequence                  int     `json:"sequence"`
	StartedAtUTC              string  `json:"started_at_utc"`
	GitCommit                 string  `json:"git_commit"`
	GoVersion                 string  `json:"go_version"`
	GOAMD64                   string  `json:"goamd64"`
	BinarySHA256              string  `json:"binary_sha256"`
	BinarySizeBytes           int64   `json:"binary_size_bytes"`
	ProfileSHA256             string  `json:"profile_sha256,omitempty"`
	PlaylistSHA256            string  `json:"playlist_sha256"`
	GuideSHA256               string  `json:"guide_sha256"`
	GuideStartUTC             string  `json:"guide_start_utc"`
	PlaylistEntries           int     `json:"playlist_entries"`
	XMLTVChannels             int     `json:"xmltv_channels"`
	XMLTVPrograms             int     `json:"xmltv_programs"`
	PlaylistImportSeconds     float64 `json:"playlist_import_seconds"`
	XEPGGenerationSeconds     float64 `json:"xepg_generation_seconds"`
	PlaylistEntriesPerSecond  float64 `json:"playlist_entries_per_second"`
	XEPGProgramsPerSecond     float64 `json:"xepg_programs_per_second"`
	StreamAttempts            int     `json:"stream_attempts"`
	StreamSuccesses           int     `json:"stream_successes"`
	StreamBytes               uint64  `json:"stream_bytes"`
	StreamBytesPerSecond      float64 `json:"stream_bytes_per_second"`
	StreamTTFBP95Milliseconds float64 `json:"stream_ttfb_p95_milliseconds"`
	ThreadfinCPUSeconds       float64 `json:"threadfin_cpu_seconds"`
	PeakRSSBytes              uint64  `json:"peak_rss_bytes"`
	TotalAllocBytes           uint64  `json:"total_alloc_bytes"`
	Mallocs                   uint64  `json:"mallocs"`
	Frees                     uint64  `json:"frees"`
	NumGC                     uint32  `json:"num_gc"`
	PauseTotalNanoseconds     uint64  `json:"pause_total_nanoseconds"`
	HeapAllocBytesAtStop      uint64  `json:"heap_alloc_bytes_at_stop"`
	ProfileStopReason         string  `json:"profile_stop_reason"`
}

type generatedArtifacts struct {
	PlaylistEntries int
	XMLTVChannels   int
	XMLTVPrograms   int
	StreamURLs      []string
}

type runtimeMetrics struct {
	StopReason            string `json:"stop_reason"`
	TotalAllocBytes       uint64 `json:"total_alloc_bytes"`
	Mallocs               uint64 `json:"mallocs"`
	Frees                 uint64 `json:"frees"`
	NumGC                 uint32 `json:"num_gc"`
	PauseTotalNanoseconds uint64 `json:"pause_total_nanoseconds"`
	HeapAllocBytesAtStop  uint64 `json:"heap_alloc_bytes_at_stop"`
}

type streamLoadResult struct {
	Attempts  int
	Successes int
	Bytes     uint64
	TTFB      []time.Duration
}

type rssSample struct {
	Peak uint64
	Err  error
}

const (
	playlistImportTimeout     = 2 * time.Minute
	guideResponseTimeout      = 2 * time.Minute
	generatedArtifactsTimeout = 3 * time.Minute
)

type processState struct {
	done       chan struct{}
	waitErr    error
	stderrPath string
}

func startProcessWait(command *exec.Cmd, stderrPath string) *processState {
	process := &processState{done: make(chan struct{}), stderrPath: stderrPath}
	go func() {
		process.waitErr = command.Wait()
		close(process.done)
	}()
	return process
}

func (process *processState) exitError(stage string) error {
	status := "exit status 0"
	if process.waitErr != nil {
		status = process.waitErr.Error()
	}
	return fmt.Errorf("Threadfin exited %s: %s; stderr: %s", stage, status, process.stderrPath)
}

func channelClosed(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	default:
		return false
	}
}

func validateProcessExit(process *processState, exitedBeforeCleanup, forced bool) error {
	if exitedBeforeCleanup {
		return process.exitError("before cleanup began")
	}
	if forced {
		return fmt.Errorf("process group required SIGKILL; stderr: %s", process.stderrPath)
	}
	var exitError *exec.ExitError
	if !errors.As(process.waitErr, &exitError) {
		return process.exitError("during cleanup with unexpected status")
	}
	waitStatus, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok || !waitStatus.Signaled() || waitStatus.Signal() != syscall.SIGTERM {
		return process.exitError("during cleanup with unexpected status")
	}
	return nil
}

func percentile95(samples []time.Duration) (time.Duration, error) {
	if len(samples) == 0 {
		return 0, errors.New("p95 requires at least one sample")
	}
	sorted := append([]time.Duration(nil), samples...)
	slices.Sort(sorted)
	rank := (95*len(sorted) + 99) / 100
	return sorted[rank-1], nil
}

func readRSS(body []byte) (uint64, error) {
	for _, line := range bytes.Split(body, []byte{'\n'}) {
		fields := bytes.Fields(line)
		if len(fields) == 3 && string(fields[0]) == "VmRSS:" && string(fields[2]) == "kB" {
			kib, err := strconv.ParseUint(string(fields[1]), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse VmRSS: %w", err)
			}
			return kib * 1024, nil
		}
	}
	return 0, errors.New("VmRSS not found")
}

func terminateProcessGroup(pgid int, processDone <-chan struct{}) (bool, error) {
	termErr := syscall.Kill(-pgid, syscall.SIGTERM)
	if termErr != nil && !errors.Is(termErr, syscall.ESRCH) {
		return false, termErr
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		parentDone := channelClosed(processDone)
		groupErr := syscall.Kill(-pgid, 0)
		if groupErr != nil && !errors.Is(groupErr, syscall.ESRCH) {
			return false, groupErr
		}
		if parentDone && errors.Is(groupErr, syscall.ESRCH) {
			return false, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	killErr := syscall.Kill(-pgid, syscall.SIGKILL)
	if killErr != nil && !errors.Is(killErr, syscall.ESRCH) {
		return true, killErr
	}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		parentDone := channelClosed(processDone)
		groupErr := syscall.Kill(-pgid, 0)
		if groupErr != nil && !errors.Is(groupErr, syscall.ESRCH) {
			return true, groupErr
		}
		if parentDone && errors.Is(groupErr, syscall.ESRCH) {
			return true, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !channelClosed(processDone) {
		return true, errors.New("Threadfin parent was not reaped after SIGKILL")
	}
	return true, errors.New("Threadfin process group still exists after SIGKILL")
}

func waitForGeneratedArtifacts(ctx context.Context, configDir, port string, process *processState) (generatedArtifacts, error) {
	deadline, cancel := context.WithTimeout(ctx, generatedArtifactsTimeout)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-process.done:
			return generatedArtifacts{}, process.exitError("before generated artifacts were ready")
		case <-deadline.Done():
			return generatedArtifacts{}, fmt.Errorf("generated artifacts: %w", deadline.Err())
		case <-ticker.C:
			got, ready, err := inspectGeneratedArtifacts(configDir, port)
			if err != nil {
				return generatedArtifacts{}, err
			}
			if ready {
				return got, nil
			}
		}
	}
}

func inspectGeneratedArtifacts(configDir, port string) (generatedArtifacts, bool, error) {
	providerPath := filepath.Join(configDir, "data", "M1000.m3u")
	provider, err := os.ReadFile(providerPath)
	if errors.Is(err, os.ErrNotExist) {
		return generatedArtifacts{}, false, nil
	}
	if err != nil {
		return generatedArtifacts{}, false, err
	}
	entries := bytes.Count(provider, []byte("#EXTINF:"))
	if entries != playlistEntryCount {
		return generatedArtifacts{}, false, nil
	}

	xmlBody, err := os.ReadFile(filepath.Join(configDir, "data", "threadfin.xml"))
	if errors.Is(err, os.ErrNotExist) {
		return generatedArtifacts{}, false, nil
	}
	if err != nil {
		return generatedArtifacts{}, false, err
	}
	var guide struct {
		Channels []struct {
			ID string `xml:"id,attr"`
		} `xml:"channel"`
		Programmes []struct {
			Channel string `xml:"channel,attr"`
		} `xml:"programme"`
	}
	if err := xml.Unmarshal(xmlBody, &guide); err != nil {
		return generatedArtifacts{}, false, nil
	}
	if len(guide.Channels) != xmltvChannelCount || len(guide.Programmes) != xmltvProgramCount {
		return generatedArtifacts{}, false, nil
	}

	m3uBody, err := os.ReadFile(filepath.Join(configDir, "data", "threadfin.m3u"))
	if errors.Is(err, os.ErrNotExist) {
		return generatedArtifacts{}, false, nil
	}
	if err != nil {
		return generatedArtifacts{}, false, err
	}
	prefix := "http://127.0.0.1:" + port + "/stream/"
	seen := make(map[string]struct{})
	urls := make([]string, 0, streamChannelCount)
	for _, line := range strings.Split(string(m3uBody), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			if _, ok := seen[line]; !ok {
				seen[line] = struct{}{}
				urls = append(urls, line)
				if len(urls) == streamChannelCount {
					break
				}
			}
		}
	}
	if len(urls) != streamChannelCount {
		return generatedArtifacts{}, false, nil
	}

	urlsBody, err := os.ReadFile(filepath.Join(configDir, "urls.json"))
	if errors.Is(err, os.ErrNotExist) {
		return generatedArtifacts{}, false, nil
	}
	if err != nil {
		return generatedArtifacts{}, false, err
	}
	var urlMap map[string]json.RawMessage
	if err := json.Unmarshal(urlsBody, &urlMap); err != nil || len(urlMap) == 0 {
		return generatedArtifacts{}, false, nil
	}
	return generatedArtifacts{entries, len(guide.Channels), len(guide.Programmes), urls}, true, nil
}

func waitForStableProvider(ctx context.Context, path string, process *processState) (time.Time, int, error) {
	deadline, cancel := context.WithTimeout(ctx, playlistImportTimeout)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	var previous int64 = -1
	for {
		select {
		case <-process.done:
			return time.Time{}, 0, process.exitError("during playlist import")
		case <-deadline.Done():
			return time.Time{}, 0, fmt.Errorf("playlist import: %w", deadline.Err())
		case now := <-ticker.C:
			info, err := os.Stat(path)
			if errors.Is(err, os.ErrNotExist) {
				previous = -1
				continue
			}
			if err != nil {
				return time.Time{}, 0, err
			}
			if info.Size() > 0 && info.Size() == previous {
				body, readErr := os.ReadFile(path)
				if errors.Is(readErr, os.ErrNotExist) {
					previous = -1
					continue
				}
				if readErr != nil {
					return time.Time{}, 0, readErr
				}
				entries := bytes.Count(body, []byte("#EXTINF:"))
				if entries == playlistEntryCount {
					return now.UTC(), entries, nil
				}
			}
			previous = info.Size()
		}
	}
}

func waitForPlaylistImport(ctx context.Context, fixture *fixtureSet, path string, process *processState) (fixtureSnapshot, time.Time, int, error) {
	deadline, cancel := context.WithTimeout(ctx, playlistImportTimeout)
	defer cancel()
	observation, err := waitFixtureSnapshot(deadline, fixture, process, func(snapshot fixtureSnapshot) bool {
		return !snapshot.PlaylistStarted.IsZero() && !snapshot.PlaylistFinished.IsZero()
	})
	if err != nil {
		return fixtureSnapshot{}, time.Time{}, 0, err
	}
	finished, entries, err := waitForStableProvider(deadline, path, process)
	if err != nil {
		return fixtureSnapshot{}, time.Time{}, 0, err
	}
	return observation, finished, entries, nil
}

func waitForProviderResponses(ctx context.Context, fixture *fixtureSet, process *processState) (fixtureSnapshot, error) {
	deadline, cancel := context.WithTimeout(ctx, guideResponseTimeout)
	defer cancel()
	observation, err := waitFixtureSnapshot(deadline, fixture, process, func(snapshot fixtureSnapshot) bool {
		return !snapshot.PlaylistFinished.IsZero() && !snapshot.GuideFinished.IsZero()
	})
	if err != nil {
		return fixtureSnapshot{}, fmt.Errorf("provider responses: %w", err)
	}
	return observation, nil
}

func signalAndReadMetrics(pid int, path string, process *processState) (runtimeMetrics, error) {
	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		if channelClosed(process.done) {
			return runtimeMetrics{}, process.exitError("before runtime metrics were requested")
		}
		return runtimeMetrics{}, err
	}
	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-process.done:
			return runtimeMetrics{}, process.exitError("before runtime metrics were published")
		case <-deadline.C:
			return runtimeMetrics{}, errors.New("runtime metrics were not published within 15 seconds")
		case <-ticker.C:
			body, err := os.ReadFile(path)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return runtimeMetrics{}, err
			}
			var metrics runtimeMetrics
			if err := json.Unmarshal(body, &metrics); err != nil {
				continue
			}
			if metrics.StopReason != "signal" {
				return runtimeMetrics{}, fmt.Errorf("profile stop reason = %q", metrics.StopReason)
			}
			return metrics, nil
		}
	}
}

func finishProcess(cmd *exec.Cmd, process *processState, cleanupStarted chan struct{}, cleanupOnce *sync.Once) (float64, bool, bool, error) {
	cleanupAlreadyStarted := channelClosed(cleanupStarted)
	exitedBeforeCleanup := !cleanupAlreadyStarted && channelClosed(process.done)
	cleanupOnce.Do(func() { close(cleanupStarted) })
	forced, err := terminateProcessGroup(cmd.Process.Pid, process.done)
	if err != nil {
		return 0, forced, false, err
	}
	cleanupSucceeded := true
	if cmd.ProcessState == nil {
		return 0, forced, cleanupSucceeded, errors.New("Threadfin process was not reaped")
	}
	cpu := cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()
	if err := validateProcessExit(process, exitedBeforeCleanup, forced); err != nil {
		return cpu.Seconds(), forced, cleanupSucceeded, err
	}
	return cpu.Seconds(), forced, cleanupSucceeded, nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeRunResult(path string, result runResult) error {
	if !filepath.IsAbs(path) {
		return errors.New("result path must be absolute")
	}
	body, err := json.Marshal(result)
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(body)
	closeErr := errors.Join(file.Sync(), file.Close())
	if err := errors.Join(writeErr, closeErr); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func validateRunConfig(config runConfig) error {
	if config.Variant != "capture" && config.Variant != "off" && config.Variant != "pgo" {
		return fmt.Errorf("invalid variant %q", config.Variant)
	}
	for name, path := range map[string]string{"binary": config.Binary, "output": config.Output} {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("%s path must be absolute", name)
		}
	}
	if config.Variant == "capture" {
		if !filepath.IsAbs(config.Profile) || config.Pair != 0 || config.Sequence < 1 || config.Sequence > 3 {
			return errors.New("capture requires an absolute profile, pair 0, and sequence 1..3")
		}
	} else if config.Profile != "" || config.Pair < 1 || config.Pair > 5 || config.Sequence < 1 || config.Sequence > 10 {
		return errors.New("measurement requires no profile, pair 1..5, and sequence 1..10")
	}
	if config.StreamDuration <= 0 || config.Clients != 32 || config.SampleBytes != 2<<20 {
		return errors.New("workload requires a positive duration, 32 clients, and 2097152 sample bytes")
	}
	_, offset := config.GuideStart.Zone()
	if config.GuideStart.IsZero() || offset != 0 {
		return errors.New("guide start must be non-zero UTC")
	}
	info, err := os.Stat(config.Binary)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&0o111 == 0 {
		return errors.New("binary is not an executable regular file")
	}
	return nil
}

func reservePort() (net.Listener, string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", err
	}
	return listener, strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), nil
}

func pilotEnvironment(metricsPath, profilePath string) []string {
	blocked := map[string]bool{"THREADFIN_PGO_METRICS": true, "THREADFIN_PGO_PROFILE": true, "THREADFIN_PGO_MAX_DURATION": true}
	environment := make([]string, 0, len(os.Environ())+3)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !blocked[key] {
			environment = append(environment, entry)
		}
	}
	environment = append(environment, "THREADFIN_PGO_METRICS="+metricsPath, "THREADFIN_PGO_MAX_DURATION=10m")
	if profilePath != "" {
		environment = append(environment, "THREADFIN_PGO_PROFILE="+profilePath)
	}
	return environment
}

func waitFixtureSnapshot(ctx context.Context, fixture *fixtureSet, process *processState, ready func(fixtureSnapshot) bool) (fixtureSnapshot, error) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		snapshot := fixture.snapshot()
		if ready(snapshot) {
			return snapshot, nil
		}
		select {
		case <-ctx.Done():
			return fixtureSnapshot{}, ctx.Err()
		case <-process.done:
			return fixtureSnapshot{}, process.exitError("before fixture observation")
		case <-ticker.C:
		}
	}
}

func samplePeakRSS(ctx context.Context, pid int, process *processState, cleanupStarted <-chan struct{}) <-chan rssSample {
	result := make(chan rssSample, 1)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		var peak uint64
		for {
			body, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
			if err == nil {
				current, parseErr := readRSS(body)
				if parseErr != nil {
					result <- rssSample{peak, parseErr}
					return
				}
				if current > peak {
					peak = current
				}
			} else if errors.Is(err, os.ErrNotExist) {
				if channelClosed(cleanupStarted) {
					result <- rssSample{Peak: peak}
					return
				}
				if channelClosed(process.done) {
					err = fmt.Errorf("%w: %v", err, process.exitError("before RSS cleanup"))
				}
				result <- rssSample{peak, err}
				return
			} else {
				result <- rssSample{peak, err}
				return
			}
			select {
			case <-ctx.Done():
				result <- rssSample{Peak: peak}
				return
			case <-process.done:
				if channelClosed(cleanupStarted) {
					result <- rssSample{Peak: peak}
				} else {
					result <- rssSample{peak, process.exitError("during RSS sampling")}
				}
				return
			case <-ticker.C:
			}
		}
	}()
	return result
}

func driveStreams(ctx context.Context, urls []string, fixture *fixtureSet, process *processState, duration time.Duration, clients int, sampleBytes int64) (streamLoadResult, error) {
	if len(urls) != streamChannelCount {
		return streamLoadResult{}, errors.New("four generated stream URLs are required")
	}
	loadCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	monitorDone := make(chan struct{})
	go func() {
		select {
		case <-process.done:
			cancel()
		case <-monitorDone:
		}
	}()
	defer close(monitorDone)
	client := &http.Client{
		Transport: &http.Transport{MaxIdleConns: clients, MaxIdleConnsPerHost: clients, IdleConnTimeout: time.Second},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()
	var result streamLoadResult
	var mu sync.Mutex
	var group sync.WaitGroup
	for worker := 0; worker < clients; worker++ {
		streamURL := urls[worker%len(urls)]
		group.Add(1)
		go func() {
			defer group.Done()
			for loadCtx.Err() == nil {
				started := time.Now()
				var firstByte time.Duration
				trace := &httptrace.ClientTrace{GotFirstResponseByte: func() { firstByte = time.Since(started) }}
				request, err := http.NewRequestWithContext(httptrace.WithClientTrace(loadCtx, trace), http.MethodGet, streamURL, nil)
				if err != nil {
					return
				}
				request.Header.Set("User-Agent", "Threadfin-PGO-Pilot/1")
				response, requestErr := client.Do(request)
				success := false
				var bytesRead int64
				if requestErr == nil {
					if response.StatusCode == http.StatusOK {
						bytesRead, requestErr = io.CopyN(io.Discard, response.Body, sampleBytes)
						success = requestErr == nil && bytesRead == sampleBytes
					}
					_ = response.Body.Close()
				}
				if loadCtx.Err() != nil && !success {
					return
				}
				mu.Lock()
				result.Attempts++
				if success {
					result.Successes++
					result.Bytes += uint64(bytesRead)
					if firstByte > 0 {
						result.TTFB = append(result.TTFB, firstByte)
					}
				}
				mu.Unlock()
			}
		}()
	}
	group.Wait()
	if channelClosed(process.done) {
		return result, process.exitError("during stream load")
	}
	if result.Attempts == 0 || 100*result.Successes < 99*result.Attempts {
		return result, fmt.Errorf("stream success = %d/%d", result.Successes, result.Attempts)
	}
	if result.Bytes < 256<<20 {
		return result, fmt.Errorf("stream bytes = %d, need at least %d", result.Bytes, uint64(256<<20))
	}
	observed := fixture.snapshot()
	for index, requests := range observed.StreamRequests {
		if requests == 0 {
			return result, fmt.Errorf("fixture stream %d was not requested", index)
		}
	}
	if len(result.TTFB) == 0 {
		return result, errors.New("no successful TTFB samples")
	}
	return result, nil
}

func bytesSHA256(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func currentCommit() (string, error) {
	command := exec.Command("git", "rev-parse", "HEAD")
	body, err := command.Output()
	if err != nil {
		return "", err
	}
	commit := strings.TrimSpace(string(body))
	if len(commit) != 40 {
		return "", fmt.Errorf("unexpected commit %q", commit)
	}
	return commit, nil
}

func binaryMetadata(path string) (string, string, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	goamd64 := ""
	for _, setting := range info.Settings {
		if setting.Key == "GOAMD64" {
			goamd64 = setting.Value
		}
	}
	if info.GoVersion == "" || goamd64 == "" {
		return "", "", errors.New("binary lacks Go version or GOAMD64 metadata")
	}
	return info.GoVersion, goamd64, nil
}

func run(ctx context.Context, config runConfig) (result runResult, err error) {
	if err := validateRunConfig(config); err != nil {
		return runResult{}, err
	}
	started := time.Now().UTC()
	outputDir := filepath.Dir(config.Output)
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return runResult{}, err
	}
	configDir, err := os.MkdirTemp("", "threadfin-pgo-run-")
	if err != nil {
		return runResult{}, err
	}
	defer os.RemoveAll(configDir)
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return runResult{}, err
	}
	fixture, err := newFixtureSet(ctx, ffmpegPath, config.GuideStart)
	if err != nil {
		return runResult{}, err
	}
	defer fixture.Close()
	listener, port, err := reservePort()
	if err != nil {
		return runResult{}, err
	}
	listenerOpen := true
	defer func() {
		if listenerOpen {
			_ = listener.Close()
		}
	}()
	settings, err := settingsDocument(fixture.baseURL(), ffmpegPath, port)
	if err != nil {
		return runResult{}, err
	}
	if err := os.WriteFile(filepath.Join(configDir, "settings.json"), settings, 0o600); err != nil {
		return runResult{}, err
	}

	stem := strings.TrimSuffix(config.Output, filepath.Ext(config.Output))
	stdoutPath := stem + ".stdout.log"
	stderrPath := stem + ".stderr.log"
	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return runResult{}, err
	}
	defer stdoutFile.Close()
	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return runResult{}, err
	}
	defer stderrFile.Close()
	metricsPath := stem + ".runtime.json"
	command := exec.Command(config.Binary, "-config="+configDir, "-port="+port, "-bind=127.0.0.1", "-debug=0")
	command.Env = pilotEnvironment(metricsPath, config.Profile)
	command.Stdout, command.Stderr = stdoutFile, stderrFile
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := listener.Close(); err != nil {
		return runResult{}, err
	}
	listenerOpen = false
	if err := command.Start(); err != nil {
		return runResult{}, err
	}
	process := startProcessWait(command, stderrPath)
	cleanupStarted := make(chan struct{})
	var cleanupOnce sync.Once
	cleanupComplete := false
	defer func() {
		if cleanupComplete {
			return
		}
		_, _, succeeded, cleanupErr := finishProcess(command, process, cleanupStarted, &cleanupOnce)
		if succeeded {
			cleanupComplete = true
		}
		if cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("cleanup: %w", cleanupErr))
		}
	}()
	rssCtx, cancelRSS := context.WithCancel(ctx)
	rssResult := samplePeakRSS(rssCtx, command.Process.Pid, process, cleanupStarted)
	defer cancelRSS()

	playlistObservation, playlistFinished, playlistEntries, err := waitForPlaylistImport(ctx, fixture, filepath.Join(configDir, "data", "M1000.m3u"), process)
	if err != nil {
		return runResult{}, err
	}
	guideObservation, err := waitForProviderResponses(ctx, fixture, process)
	if err != nil {
		return runResult{}, err
	}
	artifacts, err := waitForGeneratedArtifacts(ctx, configDir, port, process)
	if err != nil {
		return runResult{}, err
	}
	xepgFinished := time.Now().UTC()
	load, err := driveStreams(ctx, artifacts.StreamURLs, fixture, process, config.StreamDuration, config.Clients, config.SampleBytes)
	if err != nil {
		return runResult{}, err
	}
	runtimeData, err := signalAndReadMetrics(command.Process.Pid, metricsPath, process)
	if err != nil {
		return runResult{}, err
	}
	if config.Variant == "capture" {
		info, statErr := os.Stat(config.Profile)
		if statErr != nil || info.Size() == 0 {
			return runResult{}, errors.New("CPU profile was not published")
		}
	}
	cpuSeconds, _, succeeded, err := finishProcess(command, process, cleanupStarted, &cleanupOnce)
	if succeeded {
		cleanupComplete = true
	}
	if err != nil {
		return runResult{}, err
	}
	cancelRSS()
	rss := <-rssResult
	if rss.Err != nil {
		return runResult{}, rss.Err
	}
	ttfb, err := percentile95(load.TTFB)
	if err != nil {
		return runResult{}, err
	}
	binaryHash, err := fileSHA256(config.Binary)
	if err != nil {
		return runResult{}, err
	}
	binaryInfo, err := os.Stat(config.Binary)
	if err != nil {
		return runResult{}, err
	}
	commit, err := currentCommit()
	if err != nil {
		return runResult{}, err
	}
	goVersion, goamd64, err := binaryMetadata(config.Binary)
	if err != nil {
		return runResult{}, err
	}
	profileHash := ""
	if config.Profile != "" {
		profileHash, err = fileSHA256(config.Profile)
		if err != nil {
			return runResult{}, err
		}
	}
	playlistDuration := playlistFinished.Sub(playlistObservation.PlaylistStarted)
	xepgDuration := xepgFinished.Sub(guideObservation.GuideFinished)
	result = runResult{
		SchemaVersion:             1,
		Variant:                   config.Variant,
		Pair:                      config.Pair,
		Sequence:                  config.Sequence,
		StartedAtUTC:              started.Format(time.RFC3339Nano),
		GitCommit:                 commit,
		GoVersion:                 goVersion,
		GOAMD64:                   goamd64,
		BinarySHA256:              binaryHash,
		BinarySizeBytes:           binaryInfo.Size(),
		ProfileSHA256:             profileHash,
		PlaylistSHA256:            bytesSHA256(fixture.playlistBytes()),
		GuideSHA256:               bytesSHA256(fixture.guideBytes()),
		GuideStartUTC:             config.GuideStart.Format(time.RFC3339),
		PlaylistEntries:           playlistEntries,
		XMLTVChannels:             artifacts.XMLTVChannels,
		XMLTVPrograms:             artifacts.XMLTVPrograms,
		PlaylistImportSeconds:     playlistDuration.Seconds(),
		XEPGGenerationSeconds:     xepgDuration.Seconds(),
		PlaylistEntriesPerSecond:  float64(playlistEntries) / playlistDuration.Seconds(),
		XEPGProgramsPerSecond:     float64(artifacts.XMLTVPrograms) / xepgDuration.Seconds(),
		StreamAttempts:            load.Attempts,
		StreamSuccesses:           load.Successes,
		StreamBytes:               load.Bytes,
		StreamBytesPerSecond:      float64(load.Bytes) / config.StreamDuration.Seconds(),
		StreamTTFBP95Milliseconds: float64(ttfb) / float64(time.Millisecond),
		ThreadfinCPUSeconds:       cpuSeconds,
		PeakRSSBytes:              rss.Peak,
		TotalAllocBytes:           runtimeData.TotalAllocBytes,
		Mallocs:                   runtimeData.Mallocs,
		Frees:                     runtimeData.Frees,
		NumGC:                     runtimeData.NumGC,
		PauseTotalNanoseconds:     runtimeData.PauseTotalNanoseconds,
		HeapAllocBytesAtStop:      runtimeData.HeapAllocBytesAtStop,
		ProfileStopReason:         runtimeData.StopReason,
	}
	if err := writeRunResult(config.Output, result); err != nil {
		return runResult{}, err
	}
	return result, nil
}
