package src

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func restorePersistentState(t *testing.T) {
	t.Helper()
	previousSystem := System
	previousSettings := Settings
	previousData := Data
	t.Cleanup(func() {
		System = previousSystem
		Settings = previousSettings
		Data = previousData
	})
}

func TestSetProviderCompatibilityReturnsSettingsPersistenceError(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{}
	Settings.Files.M3U = map[string]interface{}{
		"provider": map[string]interface{}{},
	}
	System.File.Settings = filepath.Join(t.TempDir(), "missing", "settings.json")

	err := setProviderCompatibility("provider", "m3u", map[string]int{"streams": 1})
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("setProviderCompatibility() error = %v, want persistence path error", err)
	}
}

func TestGetGuideNumberPMSReturnsPersistenceErrorAndRollsBackID(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.PMS = make(map[string]string)
	System.File.PMS = filepath.Join(t.TempDir(), "missing", "pms.json")

	id, err := getGuideNumberPMS("channel")
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("getGuideNumberPMS() error = %v, want persistence path error", err)
	}
	if id != "" {
		t.Fatalf("getGuideNumberPMS() ID = %q after persistence failure, want empty", id)
	}
	if _, ok := Data.Cache.PMS["channel"]; ok {
		t.Fatal("getGuideNumberPMS() retained an unpersisted channel ID")
	}
}

func TestGetLineupReturnsStreamingURLPersistenceError(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{}
	Data.Cache.StreamingURLS = make(map[string]StreamInfo)
	System.File.URLS = filepath.Join(t.TempDir(), "missing", "urls.json")

	_, err := getLineup()
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("getLineup() error = %v, want persistence path error", err)
	}
}

func TestCreateXEPGDatabaseReturnsStreamingURLPersistenceError(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.StreamingURLS = make(map[string]StreamInfo)
	System.File.URLS = filepath.Join(t.TempDir(), "missing", "urls.json")

	err := createXEPGDatabase()
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("createXEPGDatabase() error = %v, want persistence path error", err)
	}
}

func TestFilterFromInterfaceReturnsEncodingError(t *testing.T) {
	_, err := filterFromInterface(map[string]interface{}{"unsupported": make(chan int)})
	if err == nil {
		t.Fatal("filterFromInterface() error = nil, want JSON encoding error")
	}
}

func TestResetStreamingURLCachePreservesMemoryOnPersistenceError(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.StreamingURLS = map[string]StreamInfo{"existing": {}}
	target := filepath.Join(t.TempDir(), "missing", "urls.json")

	err := resetStreamingURLCache(target)
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("resetStreamingURLCache() error = %v, want persistence path error", err)
	}
	if _, ok := Data.Cache.StreamingURLS["existing"]; !ok {
		t.Fatal("resetStreamingURLCache() cleared memory after persistence failure")
	}
}
