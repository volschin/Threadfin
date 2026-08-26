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
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("setProviderCompatibility() error = %v, want persistence path error", err)
	}
}

func TestGetGuideNumberPMSReturnsPersistenceErrorAndRollsBackID(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.PMS = make(map[string]string)
	System.File.PMS = filepath.Join(t.TempDir(), "missing", "pms.json")

	id, err := getGuideNumberPMS("channel")
	if _, ok := errors.AsType[*os.PathError](err); !ok {
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
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("getLineup() error = %v, want persistence path error", err)
	}
}

func TestCreateXEPGDatabaseReturnsStreamingURLPersistenceError(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.StreamingURLS = make(map[string]StreamInfo)
	System.File.URLS = filepath.Join(t.TempDir(), "missing", "urls.json")

	err := createXEPGDatabase()
	if _, ok := errors.AsType[*os.PathError](err); !ok {
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
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("resetStreamingURLCache() error = %v, want persistence path error", err)
	}
	if _, ok := Data.Cache.StreamingURLS["existing"]; !ok {
		t.Fatal("resetStreamingURLCache() cleared memory after persistence failure")
	}
}

func TestCreateM3UFileReturnsStreamingURLPersistenceError(t *testing.T) {
	restorePersistentState(t)
	System.Folder.Data = t.TempDir() + string(os.PathSeparator)
	System.File.URLS = filepath.Join(t.TempDir(), "missing", "urls.json")
	Data.XEPG.Channels = make(map[string]interface{})
	Data.Cache.StreamingURLS = make(map[string]StreamInfo)

	err := createM3UFile()
	if _, ok := errors.AsType[*os.PathError](err); !ok {
		t.Fatalf("createM3UFile() error = %v, want persistence path error", err)
	}
}

func TestSetValueForUUIDPreservesLoadError(t *testing.T) {
	restorePersistentState(t)
	System.File.XEPG = filepath.Join(t.TempDir(), "missing.json")

	err := setValueForUUID()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("setValueForUUID() error = %v, want os.ErrNotExist", err)
	}
	if _, statErr := os.Stat(System.File.XEPG); !os.IsNotExist(statErr) {
		t.Fatalf("setValueForUUID() created a replacement after load failure: %v", statErr)
	}
}

func TestGetLocalXMLTVReturnsMalformedDocumentError(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.XMLTV = make(map[string]XMLTV)
	file := filepath.Join(t.TempDir(), "malformed.xml")
	if err := os.WriteFile(file, []byte("<tv><channel>"), 0600); err != nil {
		t.Fatal(err)
	}

	var document XMLTV
	if err := getLocalXMLTV(file, &document); err == nil {
		t.Fatal("getLocalXMLTV() error = nil, want malformed XML error")
	}
}
