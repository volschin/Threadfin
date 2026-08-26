//go:build linux && amd64 && pgopilot

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"
)

var cpuProfileMu sync.Mutex

type pgoPilotConfig struct {
	ProfilePath string
	MetricsPath string
	MaxDuration time.Duration
}

type pgoPilotMetrics struct {
	StartedAtUTC          string `json:"started_at_utc"`
	StoppedAtUTC          string `json:"stopped_at_utc"`
	StopReason            string `json:"stop_reason"`
	DurationNanoseconds   int64  `json:"duration_nanoseconds"`
	TotalAllocBytes       uint64 `json:"total_alloc_bytes"`
	Mallocs               uint64 `json:"mallocs"`
	Frees                 uint64 `json:"frees"`
	NumGC                 uint32 `json:"num_gc"`
	PauseTotalNanoseconds uint64 `json:"pause_total_nanoseconds"`
	HeapAllocBytesAtStop  uint64 `json:"heap_alloc_bytes_at_stop"`
	HeapSysBytesAtStop    uint64 `json:"heap_sys_bytes_at_stop"`
}

func init() {
	metricsPath := os.Getenv("THREADFIN_PGO_METRICS")
	if metricsPath == "" {
		return
	}
	maxDuration, err := time.ParseDuration(os.Getenv("THREADFIN_PGO_MAX_DURATION"))
	if err != nil || maxDuration <= 0 {
		fmt.Fprintln(os.Stderr, "PGO pilot: THREADFIN_PGO_MAX_DURATION must be a positive duration")
		os.Exit(2)
	}
	stop := make(chan string, 1)
	signals := make(chan os.Signal, 1)
	finished := make(chan struct{})
	signal.Notify(signals, syscall.SIGUSR1)
	go func() {
		defer signal.Stop(signals)
		select {
		case <-signals:
			stop <- "signal"
		case <-finished:
		}
	}()
	done, err := startPGOPilot(pgoPilotConfig{
		ProfilePath: os.Getenv("THREADFIN_PGO_PROFILE"),
		MetricsPath: metricsPath,
		MaxDuration: maxDuration,
	}, stop)
	if err != nil {
		signal.Stop(signals)
		close(finished)
		fmt.Fprintln(os.Stderr, "PGO pilot:", err)
		os.Exit(2)
	}
	go func() {
		err := <-done
		close(finished)
		if err != nil {
			fmt.Fprintln(os.Stderr, "PGO pilot:", err)
		}
	}()
}

func startPGOPilot(config pgoPilotConfig, stop <-chan string) (<-chan error, error) {
	if !filepath.IsAbs(config.MetricsPath) {
		return nil, errors.New("PGO metrics path must be absolute")
	}
	if config.ProfilePath != "" && !filepath.IsAbs(config.ProfilePath) {
		return nil, errors.New("PGO profile path must be absolute")
	}
	if config.MaxDuration <= 0 {
		return nil, errors.New("PGO max duration must be positive")
	}

	started := time.Now().UTC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	var profile *os.File
	profileTmp := ""
	if config.ProfilePath != "" {
		profileTmp = config.ProfilePath + ".tmp"
		if err := os.MkdirAll(filepath.Dir(config.ProfilePath), 0o700); err != nil {
			return nil, err
		}
		var err error
		profile, err = os.OpenFile(profileTmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, err
		}
		cpuProfileMu.Lock()
		if err := pprof.StartCPUProfile(profile); err != nil {
			cpuProfileMu.Unlock()
			_ = profile.Close()
			_ = os.Remove(profileTmp)
			return nil, err
		}
	}

	done := make(chan error, 1)
	go func() {
		reason := "timeout"
		timer := time.NewTimer(config.MaxDuration)
		defer timer.Stop()
		select {
		case supplied := <-stop:
			if supplied != "" {
				reason = supplied
			}
		case <-timer.C:
		}

		var finishErr error
		if profile != nil {
			pprof.StopCPUProfile()
			cpuProfileMu.Unlock()
			finishErr = errors.Join(profile.Sync(), profile.Close())
			if finishErr == nil {
				finishErr = os.Rename(profileTmp, config.ProfilePath)
			}
			if finishErr != nil {
				_ = os.Remove(profileTmp)
			}
		}

		stopped := time.Now().UTC()
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		metrics := pgoPilotMetrics{
			StartedAtUTC:          started.Format(time.RFC3339Nano),
			StoppedAtUTC:          stopped.Format(time.RFC3339Nano),
			StopReason:            reason,
			DurationNanoseconds:   stopped.Sub(started).Nanoseconds(),
			TotalAllocBytes:       after.TotalAlloc - before.TotalAlloc,
			Mallocs:               after.Mallocs - before.Mallocs,
			Frees:                 after.Frees - before.Frees,
			NumGC:                 after.NumGC - before.NumGC,
			PauseTotalNanoseconds: after.PauseTotalNs - before.PauseTotalNs,
			HeapAllocBytesAtStop:  after.HeapAlloc,
			HeapSysBytesAtStop:    after.HeapSys,
		}
		finishErr = errors.Join(finishErr, writePilotJSON(config.MetricsPath, metrics))
		done <- finishErr
	}()
	return done, nil
}

func writePilotJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
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
