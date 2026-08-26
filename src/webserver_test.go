package src

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type failingJSONReader struct {
	err error
}

func (reader failingJSONReader) Read([]byte) (int, error) {
	return 0, reader.err
}

var errAPIRead = errors.New("API body read failed")
var errAPIClose = errors.New("API body close failed")

type apiBody struct {
	io.Reader
	closeErr error
	closes   int
}

func (body *apiBody) Close() error {
	body.closes++
	return body.closeErr
}

type apiReadFailure struct{}

func (apiReadFailure) Read([]byte) (int, error) {
	return 0, errAPIRead
}

func TestDecodeAPIRequestPreservesCurrentJSONSemantics(t *testing.T) {
	request, err := decodeAPIRequest(strings.NewReader(
		`{"CMD":"status","username":"benchmark","unknown":"ignored"}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if request.Cmd != "status" || request.Username != "benchmark" {
		t.Fatalf("decoded request = %#v", request)
	}

	for _, body := range []string{"", `{"cmd":`, `{"cmd":"status"} trailing`} {
		if _, err := decodeAPIRequest(strings.NewReader(body)); err == nil {
			t.Fatalf("decodeAPIRequest(%q) error = nil", body)
		}
	}

	readErr := errors.New("request read failed")
	if _, err := decodeAPIRequest(failingJSONReader{err: readErr}); !errors.Is(err, readErr) {
		t.Fatalf("decodeAPIRequest() error = %v, want %v", err, readErr)
	}
}

func TestDecodeAndCloseAPIRequestKeepsErrorsSeparate(t *testing.T) {
	body := &apiBody{Reader: apiReadFailure{}, closeErr: errAPIClose}
	_, decodeErr, closeErr := decodeAndCloseAPIRequest(body)
	if !errors.Is(decodeErr, errAPIRead) || !errors.Is(closeErr, errAPIClose) || body.closes != 1 {
		t.Fatalf("decode=%v close=%v calls=%d", decodeErr, closeErr, body.closes)
	}
}

func TestAPICloseFailureDoesNotReplaceSuccessfulStatus(t *testing.T) {
	restorePersistentState(t)
	Settings.API = true
	body := &apiBody{Reader: strings.NewReader(`{"cmd":"status"}`), closeErr: errAPIClose}
	r := httptest.NewRequest(http.MethodPost, "/api/", nil)
	r.Body = body
	w := httptest.NewRecorder()
	API(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body.closes != 1 {
		t.Fatalf("Close calls = %d, want 1", body.closes)
	}
}
