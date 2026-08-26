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

func testImageHandlerContract(t *testing.T, prefix string, handler http.HandlerFunc, configure func(string)) {
	t.Helper()
	root := t.TempDir()
	payload := bytes.Repeat([]byte{0x89, 'P', 'N', 'G'}, 256)
	if err := os.WriteFile(filepath.Join(root, "logo.png"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	configure(root + string(os.PathSeparator))
	server := httptest.NewServer(handler)
	defer server.Close()
	cases := []struct {
		name, method string
		headers      map[string]string
		wantBody     []byte
	}{
		{"GET", http.MethodGet, nil, payload},
		{"HEAD", http.MethodHead, nil, []byte{}},
		{"Range ignored", http.MethodGet, map[string]string{"Range": "bytes=1-4"}, payload},
		{"conditional ignored", http.MethodGet, map[string]string{"If-Modified-Since": "Wed, 21 Oct 2015 07:28:00 GMT"}, payload},
		{"POST", http.MethodPost, nil, payload},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, server.URL+prefix+"logo.png", nil)
			if err != nil {
				t.Fatal(err)
			}
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, readErr := io.ReadAll(resp.Body)
			if err = errors.Join(readErr, resp.Body.Close()); err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status=%d", resp.StatusCode)
			}
			if got := resp.Header.Get("Content-Type"); got != "image/png" {
				t.Fatalf("Content-Type=%q", got)
			}
			if got := resp.Header.Get("Content-Length"); got != "1024" {
				t.Fatalf("Content-Length=%q", got)
			}
			if !bytes.Equal(body, tc.wantBody) {
				t.Fatalf("body length=%d", len(body))
			}
		})
	}
	resp, err := http.Get(server.URL + prefix + "missing.png")
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(resp.Body)
	if err = errors.Join(readErr, resp.Body.Close()); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound || string(body) != "Not Found [404]\n" {
		t.Fatalf("missing=%d %q", resp.StatusCode, body)
	}
}

func TestImagesContract(t *testing.T) {
	restorePersistentState(t)
	testImageHandlerContract(t, "/images/", Images, func(path string) { System.Folder.ImagesCache = path })
}

func TestDataImagesContract(t *testing.T) {
	restorePersistentState(t)
	testImageHandlerContract(t, "/data_images/", DataImages, func(path string) { System.Folder.ImagesUpload = path })
}

func TestCopyAndCloseImageJoinsErrors(t *testing.T) {
	readErr := errors.New("read")
	writeErr := errors.New("write")
	closeErr := errors.New("close")
	reader := readCloserFunc{read: func([]byte) (int, error) { return 0, readErr }, close: func() error { return closeErr }}
	if err := copyAndCloseImage(io.Discard, reader); !errors.Is(err, readErr) || !errors.Is(err, closeErr) {
		t.Fatal(err)
	}
	reader = readCloserFunc{read: strings.NewReader("image").Read, close: func() error { return closeErr }}
	if err := copyAndCloseImage(imageWriterFunc(func([]byte) (int, error) { return 0, writeErr }), reader); !errors.Is(err, writeErr) || !errors.Is(err, closeErr) {
		t.Fatal(err)
	}
}

type readCloserFunc struct {
	read  func([]byte) (int, error)
	close func() error
}

func (r readCloserFunc) Read(p []byte) (int, error) { return r.read(p) }
func (r readCloserFunc) Close() error               { return r.close() }

type imageWriterFunc func([]byte) (int, error)

func (w imageWriterFunc) Write(p []byte) (int, error) { return w(p) }

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
