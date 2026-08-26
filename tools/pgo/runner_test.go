//go:build linux && amd64

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReadRSS(t *testing.T) {
	got, err := readRSS([]byte("Name:\tthreadfin\nState:\tS (sleeping)\nVmRSS:\t  12345 kB\nThreads:\t9\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got != 12345*1024 {
		t.Fatalf("RSS = %d", got)
	}
	if _, err := readRSS([]byte("Name:\tthreadfin\n")); err == nil {
		t.Fatal("missing VmRSS accepted")
	}
	if _, err := readRSS([]byte("VmRSS:\tnot-a-number kB\n")); err == nil {
		t.Fatal("bad VmRSS accepted")
	}
}

func TestSamplePeakRSSRejectsMissingStatusBeforeCleanup(t *testing.T) {
	process := &processState{done: make(chan struct{}), stderrPath: filepath.Join(t.TempDir(), "stderr.log")}
	cleanupStarted := make(chan struct{})
	sample := <-samplePeakRSS(context.Background(), 1<<30, process, cleanupStarted)
	if !errors.Is(sample.Err, os.ErrNotExist) {
		t.Fatalf("missing status error = %v", sample.Err)
	}
}

func TestSamplePeakRSSAllowsMissingStatusAfterCleanupStarts(t *testing.T) {
	process := &processState{done: make(chan struct{}), stderrPath: filepath.Join(t.TempDir(), "stderr.log")}
	cleanupStarted := make(chan struct{})
	close(cleanupStarted)
	sample := <-samplePeakRSS(context.Background(), 1<<30, process, cleanupStarted)
	if sample.Err != nil {
		t.Fatalf("missing status after cleanup = %v", sample.Err)
	}
}

func TestEarlyProcessExitIncludesStatusAndStderrPath(t *testing.T) {
	stderrPath := filepath.Join(t.TempDir(), "persisted.stderr.log")
	process := &processState{done: make(chan struct{}), waitErr: errors.New("exit status 23"), stderrPath: stderrPath}
	close(process.done)
	_, err := waitFixtureSnapshot(context.Background(), &fixtureSet{}, process, func(fixtureSnapshot) bool { return false })
	if err == nil {
		t.Fatal("early exit accepted")
	}
	for _, want := range []string{"exit status 23", stderrPath} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err, want)
		}
	}
}

func TestSIGTERMIsAcceptedOnlyAfterCleanupStarts(t *testing.T) {
	if os.Getenv("THREADFIN_PGO_SIGTERM_HELPER") == "1" {
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
		os.Exit(3)
	}
	command := exec.Command(os.Args[0], "-test.run=^TestSIGTERMIsAcceptedOnlyAfterCleanupStarts$")
	command.Env = append(os.Environ(), "THREADFIN_PGO_SIGTERM_HELPER=1")
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	process := startProcessWait(command, filepath.Join(t.TempDir(), "stderr.log"))
	<-process.done
	if err := validateProcessExit(process, true, false); err == nil {
		t.Fatal("SIGTERM before cleanup accepted")
	}
	if err := validateProcessExit(process, false, false); err != nil {
		t.Fatalf("SIGTERM after cleanup rejected: %v", err)
	}
}

func TestPercentile95(t *testing.T) {
	for _, tc := range []struct{ n, want int }{{1, 1}, {20, 19}, {100, 95}} {
		samples := make([]time.Duration, tc.n)
		for i := range samples {
			samples[i] = time.Duration(i+1) * time.Millisecond
		}
		got, err := percentile95(samples)
		if err != nil {
			t.Fatal(err)
		}
		if got != time.Duration(tc.want)*time.Millisecond {
			t.Fatalf("n=%d: %s", tc.n, got)
		}
	}
	if _, err := percentile95(nil); err == nil {
		t.Fatal("empty samples accepted")
	}
}

func TestTerminateProcessGroupKillsDescendantAndWaitsOnce(t *testing.T) {
	if os.Getenv("THREADFIN_PGO_HELPER") == "1" {
		child := exec.Command("sleep", "60")
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("THREADFIN_PGO_CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(2)
		}
		_ = child.Wait()
		os.Exit(0)
	}
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	cmd := exec.Command(os.Args[0], "-test.run=^TestTerminateProcessGroupKillsDescendantAndWaitsOnce$")
	cmd.Env = append(os.Environ(), "THREADFIN_PGO_HELPER=1", "THREADFIN_PGO_CHILD_PID="+pidFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	t.Cleanup(func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	})
	var childPID int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(pidFile)
		if err == nil {
			childPID, _ = strconv.Atoi(string(body))
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if childPID == 0 {
		t.Fatal("helper did not publish child PID")
	}
	forced, err := terminateProcessGroup(cmd.Process.Pid, done)
	if err != nil {
		t.Fatal(err)
	}
	if forced {
		t.Fatal("SIGKILL was required")
	}
	select {
	case <-done:
	default:
		t.Fatal("parent was not reaped")
	}
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); {
		if err := syscall.Kill(childPID, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("descendant %d still exists", childPID)
}

func TestTerminateProcessGroupForcedKillWaitsForParentAndGroup(t *testing.T) {
	if os.Getenv("THREADFIN_PGO_FORCED_HELPER") == "1" {
		signal.Ignore(syscall.SIGTERM)
		child := exec.Command("sleep", "60")
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("THREADFIN_PGO_CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0o600); err != nil {
			os.Exit(2)
		}
		_ = child.Wait()
		os.Exit(0)
	}
	pidFile := filepath.Join(t.TempDir(), "child.pid")
	cmd := exec.Command(os.Args[0], "-test.run=^TestTerminateProcessGroupForcedKillWaitsForParentAndGroup$")
	cmd.Env = append(os.Environ(), "THREADFIN_PGO_FORCED_HELPER=1", "THREADFIN_PGO_CHILD_PID="+pidFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	t.Cleanup(func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(2 * time.Second):
		}
	})
	var childPID int
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(pidFile)
		if err == nil {
			childPID, _ = strconv.Atoi(string(body))
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if childPID == 0 {
		t.Fatal("helper did not publish child PID")
	}
	forced, err := terminateProcessGroup(cmd.Process.Pid, done)
	if err != nil {
		t.Fatal(err)
	}
	if !forced {
		t.Fatal("SIGKILL was not required")
	}
	select {
	case <-done:
	default:
		t.Fatal("parent was not reaped")
	}
	if err := syscall.Kill(-cmd.Process.Pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("process group still exists: %v", err)
	}
	if err := syscall.Kill(childPID, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("descendant %d still exists: %v", childPID, err)
	}
}

func TestWaitForStableProviderRequiresTwoValidatedSamples(t *testing.T) {
	path := filepath.Join(t.TempDir(), "M1000.m3u")
	provider := "#EXTM3U\n" + strings.Repeat("#EXTINF:-1,fixture\nhttp://fixture.invalid/stream/0.ts\n", playlistEntryCount)
	if err := os.WriteFile(path, []byte(provider), 0o600); err != nil {
		t.Fatal(err)
	}
	process := runningProcessState(t)
	started := time.Now()
	_, entries, err := waitForStableProvider(context.Background(), path, process)
	if err != nil {
		t.Fatal(err)
	}
	if entries != playlistEntryCount {
		t.Fatalf("entries = %d", entries)
	}
	if elapsed := time.Since(started); elapsed < 350*time.Millisecond {
		t.Fatalf("provider accepted without two 200ms samples after %s", elapsed)
	}
}

func TestWaitForStableProviderRejectsStableWrongEntryCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "M1000.m3u")
	provider := "#EXTM3U\n" + strings.Repeat("#EXTINF:-1,fixture\nhttp://fixture.invalid/stream/0.ts\n", playlistEntryCount-1)
	if err := os.WriteFile(path, []byte(provider), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 550*time.Millisecond)
	defer cancel()
	if _, _, err := waitForStableProvider(ctx, path, runningProcessState(t)); err == nil {
		t.Fatal("stable provider with wrong entry count accepted")
	}
}

func TestWaitForPlaylistImportRequiresFixtureFinish(t *testing.T) {
	fixture := &fixtureSet{}
	fixture.observed.PlaylistStarted = time.Now().UTC()
	path := filepath.Join(t.TempDir(), "M1000.m3u")
	provider := "#EXTM3U\n" + strings.Repeat("#EXTINF:-1,fixture\nhttp://fixture.invalid/stream/0.ts\n", playlistEntryCount)
	if err := os.WriteFile(path, []byte(provider), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, _, _, err := waitForPlaylistImport(ctx, fixture, path, runningProcessState(t)); err == nil {
		t.Fatal("playlist import accepted before fixture response finished")
	}
}

func TestGeneratedReadinessRequiresBothProviderResponses(t *testing.T) {
	fixture := &fixtureSet{}
	fixture.observed.GuideFinished = time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := waitForProviderResponses(ctx, fixture, runningProcessState(t)); err == nil {
		t.Fatal("generated readiness accepted before playlist response finished")
	}
}

func TestReadinessDeadlinesRemainIndependent(t *testing.T) {
	if playlistImportTimeout != 2*time.Minute {
		t.Fatalf("playlist deadline = %s", playlistImportTimeout)
	}
	if guideResponseTimeout != 2*time.Minute {
		t.Fatalf("guide response deadline = %s", guideResponseTimeout)
	}
	if generatedArtifactsTimeout != 3*time.Minute {
		t.Fatalf("generated artifact deadline = %s", generatedArtifactsTimeout)
	}
}

func TestWriteRunResultDoesNotPublishOnParentError(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parent, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(parent, "result.json")
	if err := writeRunResult(output, runResult{SchemaVersion: 1}); err == nil {
		t.Fatal("write succeeded")
	}
	if _, err := os.Stat(output); err == nil {
		t.Fatal("final output exists")
	}
}

func TestInspectGeneratedArtifactsRejectsMalformedFiles(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "M1000.m3u"), []byte("#EXTM3U\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, ready, err := inspectGeneratedArtifacts(dir, "34401")
	if err != nil {
		t.Fatal(err)
	}
	if ready || got.PlaylistEntries != 0 || len(got.StreamURLs) != 0 {
		t.Fatalf("malformed files accepted: %+v", got)
	}
}

func TestInspectGeneratedArtifactsRejectsInvalidGeneratedFiles(t *testing.T) {
	const port = "34401"
	validStreams := generatedStreamPlaylist(port)
	for _, tc := range []struct {
		name       string
		channels   int
		programmes int
		xmlBody    []byte
		m3uBody    string
		urlsBody   string
	}{
		{name: "malformed XML", channels: xmltvChannelCount, programmes: xmltvProgramCount, xmlBody: []byte("<tv><channel"), m3uBody: validStreams, urlsBody: `{"one":{}}`},
		{name: "99 channels", channels: xmltvChannelCount - 1, programmes: xmltvProgramCount, m3uBody: validStreams, urlsBody: `{"one":{}}`},
		{name: "9999 programmes", channels: xmltvChannelCount, programmes: xmltvProgramCount - 1, m3uBody: validStreams, urlsBody: `{"one":{}}`},
		{name: "source fixture URLs", channels: xmltvChannelCount, programmes: xmltvProgramCount, m3uBody: "#EXTM3U\nhttp://fixture.invalid/stream/0.ts\nhttp://fixture.invalid/stream/1.ts\nhttp://fixture.invalid/stream/2.ts\nhttp://fixture.invalid/stream/3.ts\n", urlsBody: `{"one":{}}`},
		{name: "empty urls", channels: xmltvChannelCount, programmes: xmltvProgramCount, m3uBody: validStreams, urlsBody: `{}`},
		{name: "malformed urls", channels: xmltvChannelCount, programmes: xmltvProgramCount, m3uBody: validStreams, urlsBody: `{`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeGeneratedArtifactFixture(t, dir, tc.channels, tc.programmes, tc.xmlBody, tc.m3uBody, tc.urlsBody)
			got, ready, err := inspectGeneratedArtifacts(dir, port)
			if err != nil {
				t.Fatal(err)
			}
			if ready {
				t.Fatalf("invalid generated files accepted: %+v", got)
			}
		})
	}
}

func TestSignalAndReadMetricsRejectsMissingOrMalformedPublication(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []byte
	}{
		{name: "missing"},
		{name: "malformed", body: []byte("{")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "runtime.json")
			if tc.body != nil {
				if err := os.WriteFile(path, tc.body, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			target := exec.Command("sleep", "60")
			if err := target.Start(); err != nil {
				t.Fatal(err)
			}
			process := startProcessWait(target, filepath.Join(t.TempDir(), "stderr.log"))
			t.Cleanup(func() {
				_ = target.Process.Kill()
				select {
				case <-process.done:
				case <-time.After(2 * time.Second):
				}
			})
			if _, err := signalAndReadMetrics(target.Process.Pid, path, process); err == nil {
				t.Fatal("invalid runtime metrics accepted")
			}
		})
	}
}

func runningProcessState(t *testing.T) *processState {
	t.Helper()
	return &processState{done: make(chan struct{}), stderrPath: filepath.Join(t.TempDir(), "stderr.log")}
}

func generatedStreamPlaylist(port string) string {
	var body strings.Builder
	body.WriteString("#EXTM3U\n")
	for index := range streamChannelCount {
		fmt.Fprintf(&body, "http://127.0.0.1:%s/stream/generated-%d\n", port, index)
	}
	return body.String()
}

func writeGeneratedArtifactFixture(t *testing.T, dir string, channels, programmes int, xmlOverride []byte, m3uBody, urlsBody string) {
	t.Helper()
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o700); err != nil {
		t.Fatal(err)
	}
	provider := "#EXTM3U\n" + strings.Repeat("#EXTINF:-1,fixture\nhttp://fixture.invalid/stream/0.ts\n", playlistEntryCount)
	if err := os.WriteFile(filepath.Join(data, "M1000.m3u"), []byte(provider), 0o600); err != nil {
		t.Fatal(err)
	}
	xmlBody := xmlOverride
	if xmlBody == nil {
		var guide strings.Builder
		guide.WriteString("<tv>")
		for index := range channels {
			fmt.Fprintf(&guide, `<channel id="channel-%d"></channel>`, index)
		}
		for index := range programmes {
			fmt.Fprintf(&guide, `<programme channel="channel-%d"></programme>`, index%channels)
		}
		guide.WriteString("</tv>")
		xmlBody = []byte(guide.String())
	}
	files := map[string][]byte{
		filepath.Join(data, "threadfin.xml"): []byte(xmlBody),
		filepath.Join(data, "threadfin.m3u"): []byte(m3uBody),
		filepath.Join(dir, "urls.json"):      []byte(urlsBody),
	}
	for path, body := range files {
		if err := os.WriteFile(path, body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
