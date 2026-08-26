//go:build linux && amd64 && pgopilot

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func burnCPU(until time.Time) {
	var n uint64
	for time.Now().Before(until) {
		n = n*1664525 + 1013904223
	}
	if n == 0 {
		panic("unreachable")
	}
}

func TestStartPGOPilotPublishesAtomicFiles(t *testing.T) {
	dir := t.TempDir()
	profile := filepath.Join(dir, "cpu.pprof")
	metricsPath := filepath.Join(dir, "runtime.json")
	stop := make(chan string, 1)

	done, err := startPGOPilot(pgoPilotConfig{
		ProfilePath: profile,
		MetricsPath: metricsPath,
		MaxDuration: 2 * time.Second,
	}, stop)
	if err != nil {
		t.Fatalf("startPGOPilot() error = %v", err)
	}
	burnCPU(time.Now().Add(50 * time.Millisecond))
	stop <- "signal"
	if err := <-done; err != nil {
		t.Fatalf("completion error = %v", err)
	}

	for _, path := range []string{profile, metricsPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("%s is empty", path)
		}
		if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
			t.Fatalf("temporary file remains for %s: %v", path, err)
		}
	}

	body, err := os.ReadFile(metricsPath)
	if err != nil {
		t.Fatal(err)
	}
	var got pgoPilotMetrics
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.StopReason != "signal" || got.DurationNanoseconds <= 0 {
		t.Fatalf("unexpected metrics: %+v", got)
	}
	if got.TotalAllocBytes == 0 || got.Mallocs == 0 {
		t.Fatalf("allocation deltas missing: %+v", got)
	}
}

func TestStartPGOPilotTimesOut(t *testing.T) {
	metricsPath := filepath.Join(t.TempDir(), "runtime.json")
	done, err := startPGOPilot(pgoPilotConfig{
		MetricsPath: metricsPath,
		MaxDuration: 20 * time.Millisecond,
	}, make(chan string))
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(metricsPath)
	if err != nil {
		t.Fatal(err)
	}
	var got pgoPilotMetrics
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.StopReason != "timeout" {
		t.Fatalf("StopReason = %q", got.StopReason)
	}
}
