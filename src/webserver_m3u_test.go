package src

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"
)

func setupM3UDeliveryTest(t *testing.T) string {
	t.Helper()
	restorePersistentState(t)

	System = SystemStruct{}
	Settings = SettingsStruct{}
	Data = DataStruct{}
	dataDirectory := t.TempDir()
	System.Folder.Data = dataDirectory + string(os.PathSeparator)
	System.ServerProtocol.XML = "http"
	System.Domain = "threadfin.test"
	return filepath.Join(dataDirectory, "threadfin.m3u")
}

func requestThreadfinM3U(t *testing.T, method, target string, header http.Header) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, nil)
	for name, values := range header {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	response := httptest.NewRecorder()
	Threadfin(response, request)
	return response
}

func useM3UAuthentication(t *testing.T, auth func(*http.Request, string) error) {
	t.Helper()

	previous := m3uURLAuth
	m3uURLAuth = auth
	t.Cleanup(func() {
		m3uURLAuth = previous
	})
}

func TestThreadfinServesCachedCompleteM3UWithoutQuery(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	const cached = "#EXTM3U\ncached playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0o600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
	if response.Code != http.StatusOK || response.Body.String() != cached {
		t.Fatalf("response = (%d, %q), want (200, %q)", response.Code, response.Body.String(), cached)
	}
}

func TestThreadfinServesCachedCompleteM3UWithUnknownQuery(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	const cached = "#EXTM3U\ncached playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0o600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u?client=regression", nil)
	if response.Code != http.StatusOK || response.Body.String() != cached {
		t.Fatalf("response = (%d, %q), want (200, %q)", response.Code, response.Body.String(), cached)
	}
}

func TestThreadfinServesCachedCompleteM3UWithValidCredentials(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	authenticated := false
	useM3UAuthentication(t, func(request *http.Request, requestType string) error {
		authenticated = true
		if requestType != "m3u" {
			t.Fatalf("request type = %q, want m3u", requestType)
		}
		if request.URL.Query().Get("username") != "playlist-user" || request.URL.Query().Get("password") != "playlist-password" {
			return errors.New("invalid M3U credentials")
		}
		return nil
	})
	const cached = "#EXTM3U\ncached playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0o600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u?username=playlist-user&password=playlist-password", nil)
	if !authenticated || response.Code != http.StatusOK || response.Body.String() != cached {
		t.Fatalf("response = (authenticated=%t, %d, %q), want (true, 200, %q)", authenticated, response.Code, response.Body.String(), cached)
	}
}

func TestThreadfinM3UAuthenticationPrecedesCacheAccess(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	authenticationErr := errors.New("invalid M3U credentials")
	authenticated := false
	useM3UAuthentication(t, func(request *http.Request, requestType string) error {
		authenticated = true
		if requestType != "m3u" || request.URL.Query().Get("password") != "wrong" {
			t.Fatalf("authentication request = (%q, %q), want (m3u, wrong)", requestType, request.URL.Query().Get("password"))
		}
		return authenticationErr
	})
	const cached = "#EXTM3U\nsecret playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0o600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u?username=playlist-user&password=wrong", nil)
	if !authenticated || response.Code != http.StatusForbidden {
		t.Fatalf("response = (authenticated=%t, %d), want (true, %d)", authenticated, response.Code, http.StatusForbidden)
	}
	if response.Body.String() == cached || response.Body.String() == "" {
		t.Fatalf("body = %q, want forbidden response without cache content", response.Body.String())
	}
}

func TestM3UAuthenticationOverrideRestoresDefault(t *testing.T) {
	t.Run("override", func(t *testing.T) {
		filename := setupM3UDeliveryTest(t)
		useM3UAuthentication(t, func(*http.Request, string) error {
			return errors.New("test authentication override")
		})
		if err := os.WriteFile(filename, []byte("#EXTM3U\ncached playlist\n"), 0o600); err != nil {
			t.Fatal(err)
		}

		response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
		if response.Code != http.StatusForbidden {
			t.Fatalf("override status = %d, want %d", response.Code, http.StatusForbidden)
		}
	})

	filename := setupM3UDeliveryTest(t)
	const cached = "#EXTM3U\ncached playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0o600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
	if response.Code != http.StatusOK || response.Body.String() != cached {
		t.Fatalf("restored default response = (%d, %q), want (200, %q)", response.Code, response.Body.String(), cached)
	}
}

func TestThreadfinBuildsAndServesMissingCompleteM3U(t *testing.T) {
	filename := setupM3UDeliveryTest(t)

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
	published, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || response.Body.String() != string(published) {
		t.Fatalf("response = (%d, %q), want published bytes (200, %q)", response.Code, response.Body.String(), published)
	}
}

func TestThreadfinCompleteM3UHTTPSemantics(t *testing.T) {
	const cachedContent = "#EXTM3U\ncached playlist with enough bytes\n"
	tests := []struct {
		name              string
		method            string
		header            http.Header
		wantStatus        int
		wantBody          func([]byte) string
		wantContentLength func([]byte) string
		wantContentRange  func([]byte) string
	}{
		{
			name:              "GET",
			method:            http.MethodGet,
			wantStatus:        http.StatusOK,
			wantBody:          func(content []byte) string { return string(content) },
			wantContentLength: func(content []byte) string { return strconv.Itoa(len(content)) },
		},
		{
			name:              "HEAD",
			method:            http.MethodHead,
			wantStatus:        http.StatusOK,
			wantBody:          func([]byte) string { return "" },
			wantContentLength: func(content []byte) string { return strconv.Itoa(len(content)) },
		},
		{
			name:              "byte range",
			method:            http.MethodGet,
			header:            http.Header{"Range": {"bytes=1-5"}},
			wantStatus:        http.StatusPartialContent,
			wantBody:          func(content []byte) string { return string(content[1:6]) },
			wantContentLength: func([]byte) string { return "5" },
			wantContentRange:  func(content []byte) string { return fmt.Sprintf("bytes 1-5/%d", len(content)) },
		},
		{
			name:       "If-Modified-Since",
			method:     http.MethodGet,
			header:     http.Header{"If-Modified-Since": {time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)}},
			wantStatus: http.StatusNotModified,
			wantBody:   func([]byte) string { return "" },
		},
	}

	for _, source := range []string{"cached", "generated"} {
		for _, test := range tests {
			t.Run(source+"/"+test.name, func(t *testing.T) {
				filename := setupM3UDeliveryTest(t)
				if source == "cached" {
					if err := os.WriteFile(filename, []byte(cachedContent), 0o600); err != nil {
						t.Fatal(err)
					}
				}

				response := requestThreadfinM3U(t, test.method, "/m3u/threadfin.m3u", test.header)
				completeContent, err := os.ReadFile(filename)
				if err != nil {
					t.Fatal(err)
				}
				if response.Code != test.wantStatus {
					t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
				}
				if got, want := response.Body.String(), test.wantBody(completeContent); got != want {
					t.Fatalf("body = %q, want %q", got, want)
				}

				if test.wantStatus != http.StatusNotModified {
					if got, want := response.Header().Get("Content-Type"), "audio/x-mpegurl"; got != want {
						t.Fatalf("Content-Type = %q, want %q", got, want)
					}
				}
				wantLength := ""
				if test.wantContentLength != nil {
					wantLength = test.wantContentLength(completeContent)
				}
				if got := response.Header().Get("Content-Length"); got != wantLength {
					t.Fatalf("Content-Length = %q, want %q", got, wantLength)
				}
				wantRange := ""
				if test.wantContentRange != nil {
					wantRange = test.wantContentRange(completeContent)
				}
				if got := response.Header().Get("Content-Range"); got != wantRange {
					t.Fatalf("Content-Range = %q, want %q", got, wantRange)
				}
				if test.wantStatus == http.StatusPartialContent && response.Header().Get("Accept-Ranges") != "bytes" {
					t.Fatalf("Accept-Ranges = %q, want bytes", response.Header().Get("Accept-Ranges"))
				}
				if response.Header().Get("Last-Modified") == "" {
					t.Fatal("Last-Modified header is empty")
				}

				wantDisposition := ""
				if source == "generated" {
					wantDisposition = "attachment; filename=threadfin.m3u"
				}
				if got := response.Header().Get("Content-Disposition"); got != wantDisposition {
					t.Fatalf("Content-Disposition = %q, want %q", got, wantDisposition)
				}
			})
		}
	}
}

func TestServeOpenedM3USetsStableContentType(t *testing.T) {
	const content = "#EXTM3U\nportable content type\n"
	file := &m3uTestFile{
		Reader: bytes.NewReader([]byte(content)),
		info: m3uTestFileInfo{
			name: "playlist.unknown",
			size: int64(len(content)),
			mode: 0o600,
		},
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil)

	if err := serveOpenedM3U(response, request, file); err != nil {
		t.Fatalf("serveOpenedM3U() error = %v", err)
	}
	if got, want := response.Header().Get("Content-Type"), "audio/x-mpegurl"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
}

func TestThreadfinReturnsServerErrorForCompleteM3UCacheDirectory(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	if err := os.Mkdir(filename, 0o700); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
	if response.Code != http.StatusInternalServerError || response.Body.String() == "" {
		t.Fatalf("response = (%d, %q), want non-empty 500", response.Code, response.Body.String())
	}
}

func TestThreadfinReturnsServerErrorForMissingM3UDataDirectory(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	System.Folder.Data = filepath.Join(filepath.Dir(filename), "missing") + string(os.PathSeparator)

	response := requestThreadfinM3U(t, http.MethodGet, "/m3u/threadfin.m3u", nil)
	if response.Code != http.StatusInternalServerError || response.Body.String() == "" {
		t.Fatalf("response = (%d, %q), want non-empty 500", response.Code, response.Body.String())
	}
}

type m3uTestFile struct {
	*bytes.Reader
	info     os.FileInfo
	statErr  error
	closeErr error
	closed   bool
}

func (f *m3uTestFile) Stat() (os.FileInfo, error) { return f.info, f.statErr }

func (f *m3uTestFile) Close() error {
	f.closed = true
	return f.closeErr
}

type m3uTestFileInfo struct {
	name string
	size int64
	mode os.FileMode
}

func (i m3uTestFileInfo) Name() string       { return i.name }
func (i m3uTestFileInfo) Size() int64        { return i.size }
func (i m3uTestFileInfo) Mode() os.FileMode  { return i.mode }
func (i m3uTestFileInfo) ModTime() time.Time { return time.Time{} }
func (i m3uTestFileInfo) IsDir() bool        { return i.mode.IsDir() }
func (i m3uTestFileInfo) Sys() interface{}   { return nil }

type blockingM3UReadFile struct {
	*bytes.Reader
	info        os.FileInfo
	readStarted chan struct{}
	releaseRead chan struct{}
	startOnce   sync.Once
	closeOnce   sync.Once
	closed      chan struct{}
}

func (f *blockingM3UReadFile) Read(p []byte) (int, error) {
	f.startOnce.Do(func() { close(f.readStarted) })
	<-f.releaseRead
	return f.Reader.Read(p)
}

func (f *blockingM3UReadFile) Stat() (os.FileInfo, error) { return f.info, nil }

func (f *blockingM3UReadFile) Close() error {
	f.closeOnce.Do(func() { close(f.closed) })
	return nil
}

func TestM3UPublicationProceedsWhileCompleteReaderActive(t *testing.T) {
	const currentContent = "#EXTM3U\ncurrent playlist\n"
	const replacementContent = "#EXTM3U\nreplacement playlist\n"
	filename := filepath.Join(t.TempDir(), "threadfin.m3u")
	releaseRead := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseRead) }) }
	t.Cleanup(release)

	activeFile := &blockingM3UReadFile{
		Reader:      bytes.NewReader([]byte(currentContent)),
		info:        m3uTestFileInfo{name: "threadfin.m3u", size: int64(len(currentContent)), mode: 0o600},
		readStarted: make(chan struct{}),
		releaseRead: releaseRead,
		closed:      make(chan struct{}),
	}
	readerResponse := httptest.NewRecorder()
	readerDone := make(chan error, 1)
	go func() {
		opened, err := openAndServeM3U(
			readerResponse,
			httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil),
			filename,
			func(string) (m3uReadFile, error) { return activeFile, nil },
		)
		if !opened && err == nil {
			err = errors.New("complete M3U reader was not opened")
		}
		readerDone <- err
	}()

	select {
	case <-activeFile.readStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("complete M3U reader did not become active")
	}

	renameBoundary := make(chan struct{})
	publicationDone := make(chan error, 1)
	go func() {
		publicationDone <- publishM3UFileWithOps(filename, func(writer io.StringWriter) error {
			_, err := writer.WriteString(replacementContent)
			return err
		}, m3uPublicationOps{
			createTemp: func(string, string) (m3uTempFile, error) {
				return &publicationTestFile{name: filename + ".tmp"}, nil
			},
			rename: func(string, string) error {
				close(renameBoundary)
				return nil
			},
			remove: func(string) error { return nil },
		})
	}()

	select {
	case <-renameBoundary:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not reach the rename boundary while reader was active")
	}
	select {
	case err := <-publicationDone:
		if err != nil {
			t.Fatalf("publishM3UFileWithOps() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("publication did not complete at the rename boundary")
	}
	select {
	case <-activeFile.closed:
		t.Fatal("reader closed before publication completed")
	default:
	}

	release()
	select {
	case err := <-readerDone:
		if err != nil {
			t.Fatalf("serveOpenedM3U() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("complete M3U reader did not finish")
	}
	if got, want := readerResponse.Body.String(), currentContent; got != want {
		t.Fatalf("reader body = %q, want %q", got, want)
	}
}

func TestServeM3UUsesRequestedGroupsInsteadOfCompleteCache(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	var builtGroups []string
	ops := m3uHandlerOps{
		build: func(groups []string) (string, error) {
			builtGroups = append([]string(nil), groups...)
			return "#EXTM3U\nfiltered playlist\n", nil
		},
		open: func(string) (m3uReadFile, error) {
			t.Fatal("filtered request attempted to open complete cache")
			return nil, nil
		},
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u?group-title=News,Sports", nil)
	if err := serveM3U(response, request, filename, "News,Sports", ops); err != nil {
		t.Fatalf("serveM3U() error = %v", err)
	}
	if !slices.Equal(builtGroups, []string{"News", "Sports"}) {
		t.Fatalf("build groups = %v", builtGroups)
	}
	if response.Body.String() != "#EXTM3U\nfiltered playlist\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestServeM3UReturnsInitialOpenErrorWithoutWritingContent(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	openErr := errors.New("complete cache unavailable")
	buildCalled := false
	ops := m3uHandlerOps{
		build: func([]string) (string, error) {
			buildCalled = true
			return "", nil
		},
		open: func(string) (m3uReadFile, error) { return nil, openErr },
	}
	response := httptest.NewRecorder()

	err := serveM3U(response, httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil), filename, "", ops)
	if !errors.Is(err, openErr) || buildCalled || response.Body.Len() != 0 {
		t.Fatalf("serveM3U() = (%v, build=%t, body=%q), want initial open error without build or content", err, buildCalled, response.Body.String())
	}
}

func TestServeM3UReturnsReopenErrorAfterBuildingWithoutWritingContent(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	reopenErr := errors.New("published cache unavailable")
	openCount := 0
	ops := m3uHandlerOps{
		build: func(groups []string) (string, error) {
			if groups != nil {
				t.Fatalf("build groups = %v, want nil", groups)
			}
			return "", nil
		},
		open: func(string) (m3uReadFile, error) {
			openCount++
			if openCount == 1 {
				return nil, fs.ErrNotExist
			}
			return nil, reopenErr
		},
	}
	response := httptest.NewRecorder()

	err := serveM3U(response, httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil), filename, "", ops)
	if !errors.Is(err, reopenErr) || openCount != 2 || response.Body.Len() != 0 {
		t.Fatalf("serveM3U() = (%v, opens=%d, body=%q), want reopen error without content", err, openCount, response.Body.String())
	}
}

func TestServeM3UReturnsFilteredBuildErrorWithoutWritingContent(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	buildErr := errors.New("filtered build failed")
	ops := m3uHandlerOps{
		build: func(groups []string) (string, error) {
			if !slices.Equal(groups, []string{"News"}) {
				t.Fatalf("build groups = %v, want [News]", groups)
			}
			return "", buildErr
		},
		open: func(string) (m3uReadFile, error) {
			t.Fatal("filtered request attempted to open complete cache")
			return nil, nil
		},
	}
	response := httptest.NewRecorder()

	err := serveM3U(response, httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u?group-title=News", nil), filename, "News", ops)
	if !errors.Is(err, buildErr) || response.Body.Len() != 0 {
		t.Fatalf("serveM3U() = (%v, body=%q), want filtered build error without content", err, response.Body.String())
	}
}

func TestServeM3UReturnsErrorForNonRegularCompleteCacheWithoutWritingContent(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	file := &m3uTestFile{Reader: bytes.NewReader(nil), info: m3uTestFileInfo{name: "threadfin.m3u", mode: os.ModeDir}}
	ops := m3uHandlerOps{
		build: func([]string) (string, error) {
			t.Fatal("existing complete cache attempted to build")
			return "", nil
		},
		open: func(string) (m3uReadFile, error) { return file, nil },
	}
	response := httptest.NewRecorder()

	err := serveM3U(response, httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil), filename, "", ops)
	if err == nil || !file.closed || response.Body.Len() != 0 {
		t.Fatalf("serveM3U() = (%v, closed=%t, body=%q), want non-regular error without content", err, file.closed, response.Body.String())
	}
}

func TestServeM3UReturnsStatErrorWithoutWritingContent(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	statErr := errors.New("stat failed")
	file := &m3uTestFile{Reader: bytes.NewReader(nil), statErr: statErr}
	ops := m3uHandlerOps{
		build: func([]string) (string, error) {
			t.Fatal("existing complete cache attempted to build")
			return "", nil
		},
		open: func(string) (m3uReadFile, error) { return file, nil },
	}
	response := httptest.NewRecorder()

	err := serveM3U(response, httptest.NewRequest(http.MethodGet, "/m3u/threadfin.m3u", nil), filename, "", ops)
	if !errors.Is(err, statErr) || !file.closed || response.Body.Len() != 0 {
		t.Fatalf("serveM3U() = (%v, closed=%t, body=%q), want stat error without content", err, file.closed, response.Body.String())
	}
}
