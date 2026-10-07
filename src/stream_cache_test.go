package src

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStreamCacheConcurrentLazyLoad(t *testing.T) {
	restorePersistentState(t)
	Data.Cache.StreamingURLS = nil
	System.File.URLS = filepath.Join(t.TempDir(), "urls.json")
	if err := os.WriteFile(System.File.URLS, []byte(`{"channel":{"url":"http://provider.test/live\r\n","name":"Channel"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var callers sync.WaitGroup
	start := make(chan struct{})
	for range 8 {
		callers.Go(func() {
			<-start
			stream, err := getStreamInfo("channel")
			if err != nil || stream.Name != "Channel" || stream.URL != "http://provider.test/live" {
				t.Errorf("unexpected stream: %#v, %v", stream, err)
			}
		})
	}
	close(start)
	callers.Wait()
}

func TestStreamCacheConcurrentPublicationAndReset(t *testing.T) {
	restorePersistentState(t)
	Settings = SettingsStruct{}
	System.File.URLS = filepath.Join(t.TempDir(), "urls.json")
	if err := resetStreamingURLCache(System.File.URLS); err != nil {
		t.Fatal(err)
	}
	var callers sync.WaitGroup
	for index := range 8 {
		callers.Go(func() {
			for iteration := range 20 {
				url := fmt.Sprintf("http://provider.test/%d/%d", index, iteration)
				if _, err := createStreamingURL("M3U", "provider", "1", "Channel", url, nil, nil, nil); err != nil {
					t.Error(err)
				}
				_, _ = getStreamInfo(getMD5("provider-" + url))
				if index == 0 {
					if err := resetStreamingURLCache(System.File.URLS); err != nil {
						t.Error(err)
					}
				} else if err := saveStreamingURLCache(System.File.URLS); err != nil {
					t.Error(err)
				}
			}
		})
	}
	callers.Wait()
	if _, err := loadJSONFileToMap(System.File.URLS); err != nil {
		t.Fatal(err)
	}
}

func TestStreamCacheDoesNotShareBackupPointers(t *testing.T) {
	restorePersistentState(t)
	clearStreamingURLCache()
	backup := &BackupStream{URL: "http://backup.test/original"}
	_, err := createStreamingURL("M3U", "provider", "1", "Channel", "http://provider.test/stream", backup, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	backup.URL = "http://modified.test/source"
	id := getMD5("provider-http://provider.test/stream")
	first, err := getStreamInfo(id)
	if err != nil {
		t.Fatal(err)
	}
	if first.BackupChannel1.URL != "http://backup.test/original" {
		t.Fatal("cache retained caller-owned pointer")
	}
	first.BackupChannel1.URL = "http://modified.test/reader"
	second, err := getStreamInfo(id)
	if err != nil {
		t.Fatal(err)
	}
	if second.BackupChannel1.URL != "http://backup.test/original" {
		t.Fatal("reader modified cached backup")
	}
}
