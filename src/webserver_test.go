package src

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestImageHandlersPreserveCurrentDeliveryContract(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "cache")
	uploadDir := filepath.Join(t.TempDir(), "upload")
	for _, dir := range []string{cacheDir, uploadDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	systemMutex.Lock()
	previousCache := System.Folder.ImagesCache
	previousUpload := System.Folder.ImagesUpload
	System.Folder.ImagesCache = cacheDir + string(os.PathSeparator)
	System.Folder.ImagesUpload = uploadDir + string(os.PathSeparator)
	systemMutex.Unlock()
	t.Cleanup(func() {
		systemMutex.Lock()
		System.Folder.ImagesCache = previousCache
		System.Folder.ImagesUpload = previousUpload
		systemMutex.Unlock()
	})

	content := []byte{0x89, 'P', 'N', 'G', 0x01, 0x02}
	cases := []struct {
		name    string
		handler http.HandlerFunc
		dir     string
		path    string
	}{
		{name: "cache", handler: Images, dir: cacheDir, path: "/images/cached.png"},
		{name: "upload", handler: DataImages, dir: uploadDir, path: "/data_images/uploaded.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Base(tc.path)
			if err := os.WriteFile(filepath.Join(tc.dir, filename), content, 0o600); err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			tc.handler(response, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if got := response.Header().Get("Content-Type"); got != "image/png" {
				t.Fatalf("Content-Type = %q, want image/png", got)
			}
			if got := response.Header().Get("Content-Length"); got != strconv.Itoa(len(content)) {
				t.Fatalf("Content-Length = %q, want %d", got, len(content))
			}
			if !bytes.Equal(response.Body.Bytes(), content) {
				t.Fatalf("body = %v, want %v", response.Body.Bytes(), content)
			}

			missing := httptest.NewRecorder()
			tc.handler(missing, httptest.NewRequest(http.MethodGet, tc.path+".missing", nil))
			if missing.Code != http.StatusNotFound {
				t.Fatalf("missing status = %d, want %d", missing.Code, http.StatusNotFound)
			}
		})
	}
}
