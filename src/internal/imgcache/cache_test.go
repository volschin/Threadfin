package imgcache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	c, err := New(t.TempDir()+string(os.PathSeparator), "http://threadfin.test/images/", true)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func imageURL(c *Cache, src string) string {
	return c.Image.GetURL(src, "", "", false, 0, "")
}

func TestCachingDoesNotBlockLookupsOrLoseNewQueueEntries(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		io.WriteString(w, "complete image")
	}))
	defer server.Close()
	defer unblock()
	c := newTestCache(t)
	src := server.URL + "/image.png?token=value"
	imageURL(c, src)
	done := make(chan struct{})
	go func() { c.Image.Caching(); close(done) }()
	<-started
	lookup := make(chan string, 1)
	other := server.URL + "/later.png"
	go func() { lookup <- imageURL(c, other) }()
	select {
	case got := <-lookup:
		if got != other {
			t.Fatalf("queued lookup = %q, want source", got)
		}
	case <-time.After(time.Second):
		unblock()
		<-done
		<-lookup
		t.Fatal("lookup blocked behind an image download")
	}
	unblock()
	<-done
	filename := strToMD5(server.URL+"/image.png") + ".png"
	if got := imageURL(c, src); got != c.cacheURL+filename {
		t.Fatalf("cached URL = %q, want %q", got, c.cacheURL+filename)
	}
	if len(c.Queue) != 1 || c.Queue[0] != other {
		t.Fatalf("remaining queue = %v, want only %s", c.Queue, other)
	}
	content, err := os.ReadFile(filepath.Join(c.path, filename))
	if err != nil || string(content) != "complete image" {
		t.Fatalf("published image = %q, error %v", content, err)
	}
}

func TestCachingRejectsOversizedAndIncompleteImages(t *testing.T) {
	for _, kind := range []string{"oversized", "incomplete"} {
		t.Run(kind, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if kind == "incomplete" {
					w.Header().Set("Content-Length", "100")
					io.WriteString(w, "partial")
					return
				}
				io.Copy(w, strings.NewReader(strings.Repeat("x", 10*1024*1024+1)))
			}))
			defer server.Close()
			c := newTestCache(t)
			src := server.URL + "/image.png"
			imageURL(c, src)
			c.Image.Caching()
			if got := imageURL(c, src); got != src {
				t.Errorf("failed image published as %q", got)
			}
			files, err := os.ReadDir(c.path)
			if err != nil || len(files) != 0 {
				t.Fatalf("failed download left %d files, error %v", len(files), err)
			}
		})
	}
}

func TestConcurrentCachingAndRemovalPublishOneCompleteImage(t *testing.T) {
	var requests atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(started)
		}
		io.WriteString(w, "first")
		w.(http.Flusher).Flush()
		<-release
		io.WriteString(w, "last")
	}))
	defer server.Close()
	c := newTestCache(t)
	src := server.URL + "/image.png?token=value"
	imageURL(c, src)
	var workers sync.WaitGroup
	workers.Go(c.Image.Caching)
	<-started
	workers.Go(c.Image.Caching)
	removed := make(chan struct{})
	go func() { c.Image.Remove(); close(removed) }()
	select {
	case <-removed:
	case <-time.After(time.Second):
		close(release)
		workers.Wait()
		<-removed
		t.Fatal("cache cleanup blocked behind a download")
	}
	close(release)
	workers.Wait()
	if requests.Load() != 1 {
		t.Fatalf("downloaded image %d times, want 1", requests.Load())
	}
	filename := strToMD5(server.URL+"/image.png") + ".png"
	content, err := os.ReadFile(filepath.Join(c.path, filename))
	if err != nil || string(content) != "firstlast" {
		t.Fatalf("published image = %q, error %v", content, err)
	}
}

func TestCachingDeadlineCoversStalledResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "partial")
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	c := newTestCache(t)
	if c.client.Timeout <= 0 || c.client.Timeout > 30*time.Second {
		t.Fatalf("image client timeout = %v, want a bounded deadline", c.client.Timeout)
	}
	c.client.Timeout = 50 * time.Millisecond
	src := server.URL + "/image.png"
	imageURL(c, src)
	c.Image.Caching()
	if got := imageURL(c, src); got != src || c.Pending() != 1 {
		t.Fatalf("timed out image = %q, pending = %d", got, c.Pending())
	}
	files, err := os.ReadDir(c.path)
	if err != nil || len(files) != 0 {
		t.Fatalf("timed out download left %d files, error %v", len(files), err)
	}
}

func TestRemovingUnusedImageUpdatesCachedInventory(t *testing.T) {
	directory := t.TempDir()
	src := "http://example.test/logo.png"
	filename := strToMD5(src) + ".png"
	if err := os.WriteFile(filepath.Join(directory, filename), []byte("unused image"), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := New(directory, "/images/", true)
	if err != nil {
		t.Fatal(err)
	}
	c.Image.Remove()
	if got := imageURL(c, src); got != src || c.Pending() != 1 {
		t.Fatalf("deleted image returned %q, pending = %d", got, c.Pending())
	}
}
