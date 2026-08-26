//go:build linux && amd64

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
			t.Cleanup(func() {
				_ = target.Process.Kill()
				_ = target.Wait()
			})
			if _, err := signalAndReadMetrics(target.Process.Pid, path); err == nil {
				t.Fatal("invalid runtime metrics accepted")
			}
		})
	}
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
