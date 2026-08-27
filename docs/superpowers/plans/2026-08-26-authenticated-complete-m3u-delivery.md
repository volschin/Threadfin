# Authenticated Complete-M3U Delivery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Serve cached or newly generated complete M3U bytes for authenticated requests while atomically publishing complete playlists and preserving filtered-playlist behavior.

**Architecture:** Separate complete-file inspection and HTTP response selection behind small dependency-injected helpers. Keep complete playlists streamed into a same-directory temporary file, then flush, sync, close, and atomically rename it; keep filtered playlists in memory.

**Tech Stack:** Go 1.27, standard `net/http`, `io`, `bufio`, `os`, and `testing` packages, checked-in vendor tree

**Spec:** `docs/superpowers/specs/2026-08-26-stream-reliability-fixes-design.md`

## Global Constraints

- Keep `urlAuth` as first M3U gate; do not change authentication methods, credentials, or authorization policy.
- Select complete versus filtered behavior only from `group-title`; credentials and unknown query parameters are non-functional.
- Keep complete generation streamed; do not build full playlist in memory.
- Publish complete playlist through same-directory temporary file, successful flush/sync/close, then `os.Rename`; never remove final cache before rename.
- Keep filtered generation in memory and preserve comma-separated `group-title` semantics.
- Preserve `buildM3U` parsing, sorting, clamping, group counts, duplicate detection, URL rewriting, and merged modernization.
- Preserve successful cached/generated headers, content type, download filename, range, HEAD, and conditional response behavior.
- Return HTTP 500 for inspection, build, publication, and post-build access errors; never return empty HTTP 200 for these failures.
- Do not use `t.Parallel`; tests mutate package and authentication globals.
- Do not change dependencies, vendor, schemas, assets, versions, release metadata, or unrelated XMLTV behavior.
- Run all tests and builds with checked-in vendor tree.

## File Structure

- Modify `src/m3u.go`: atomic complete-playlist publication and shared renderer for complete/filtered output.
- Create `src/m3u_publication_test.go`: publication transaction and failure-preservation tests.
- Modify `src/webserver.go`: complete/filtered selection, opened-file serving, and HTTP error integration.
- Create `src/webserver_m3u_test.go`: handler-level authentication/cache/build/error tests and dependency-seam tests.
- Read-only context: `src/authentication.go`, `src/internal/authentication`, `src/persistence_test.go`, and `src/xepg.go:createM3UFile`.

---

### Task 1: Add atomic complete-playlist publication transaction

**Files:**
- Modify: `src/m3u.go:3-17`
- Modify: `src/m3u.go:233-399`
- Create: `src/m3u_publication_test.go`

**Interfaces:**
- Consumes: existing `buildM3U(groups []string) (string, error)` output loop.
- Produces: `publishM3UFile(filename string, write func(io.StringWriter) error) error`; injectable `publishM3UFileWithOps`; `m3uTempFile`; `m3uPublicationOps`.

- [ ] **Step 1: Define failing publication test doubles**

```go
package src

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type publicationTestFile struct {
	bytes.Buffer
	name     string
	writeErr error
	syncErr  error
	closeErr error
}

func (f *publicationTestFile) Name() string { return f.name }
func (f *publicationTestFile) Write(p []byte) (int, error) {
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return f.Buffer.Write(p)
}
func (f *publicationTestFile) Sync() error  { return f.syncErr }
func (f *publicationTestFile) Close() error { return f.closeErr }
```

- [ ] **Step 2: Write failing successful-publication test**

```go
func TestPublishM3UFilePublishesCompleteFile(t *testing.T) {
	directory := t.TempDir()
	filename := filepath.Join(directory, "threadfin.m3u")

	err := publishM3UFile(filename, func(writer io.StringWriter) error {
		_, err := writer.WriteString("#EXTM3U\nnew playlist\n")
		return err
	})
	if err != nil {
		t.Fatalf("publishM3UFile() error = %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "#EXTM3U\nnew playlist\n"; got != want {
		t.Fatalf("published content = %q, want %q", got, want)
	}
	temporary, err := filepath.Glob(filepath.Join(directory, ".threadfin.m3u-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporary) != 0 {
		t.Fatalf("temporary files remain after publication: %v", temporary)
	}
}
```

- [ ] **Step 3: Write table-driven failure-preservation test**

Cover writer error, buffered-flush error, sync error, close error, and rename error. For each case, create existing final file containing `published playlist`, return an injected temporary file, assert returned error wraps expected sentinel, assert final bytes remain unchanged, and assert temporary path is removed.

```go
type m3uPublicationOps struct {
	createTemp func(string, string) (m3uTempFile, error)
	rename     func(string, string) error
	remove     func(string) error
}
```

Use a small write for flush case so `bufio.Writer` defers injected `writeErr` until `Flush`; use `strings.Repeat("x", (1<<20)+1)` for immediate writer error. Add a create-temp error case and a remove-error case that asserts `errors.Join` preserves both primary and cleanup errors.

- [ ] **Step 4: Run publication tests and verify RED**

```sh
go test -count=1 -mod=vendor ./src -run '^TestPublishM3UFile'
```

Expected: compile failure because publication interfaces and functions do not exist.

- [ ] **Step 5: Implement publication interfaces and transaction**

Add imports `errors`, `io`, `io/fs`, and `path/filepath` while retaining `bufio`.

```go
type m3uTempFile interface {
	io.Writer
	Name() string
	Sync() error
	Close() error
}

type m3uPublicationOps struct {
	createTemp func(string, string) (m3uTempFile, error)
	rename     func(string, string) error
	remove     func(string) error
}

func publishM3UFile(filename string, write func(io.StringWriter) error) error {
	return publishM3UFileWithOps(filename, write, m3uPublicationOps{
		createTemp: func(directory, pattern string) (m3uTempFile, error) {
			return os.CreateTemp(directory, pattern)
		},
		rename: os.Rename,
		remove: os.Remove,
	})
}

func publishM3UFileWithOps(filename string, write func(io.StringWriter) error, ops m3uPublicationOps) (err error) {
	temporary, err := ops.createTemp(filepath.Dir(filename), ".threadfin.m3u-*")
	if err != nil {
		return fmt.Errorf("create temporary M3U file: %w", err)
	}

	temporaryName := temporary.Name()
	closed := false
	published := false
	defer func() {
		if !closed {
			err = errors.Join(err, temporary.Close())
		}
		if !published {
			removeErr := ops.remove(temporaryName)
			if removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
				err = errors.Join(err, fmt.Errorf("remove temporary M3U file: %w", removeErr))
			}
		}
	}()

	writer := bufio.NewWriterSize(temporary, 1<<20)
	if err := write(writer); err != nil {
		return fmt.Errorf("write complete M3U: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush complete M3U: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync complete M3U: %w", err)
	}
	closed = true
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close complete M3U: %w", err)
	}
	if err := ops.rename(temporaryName, filename); err != nil {
		return fmt.Errorf("publish complete M3U: %w", err)
	}
	published = true
	return nil
}
```

- [ ] **Step 6: Run publication tests and verify GREEN**

Run same command. Expected: PASS.

- [ ] **Step 7: Commit task only when explicitly requested**

```sh
git add src/m3u.go src/m3u_publication_test.go
git commit -m "fix: publish complete M3U atomically"
```

---

### Task 2: Route complete and filtered `buildM3U` output through shared renderer

**Files:**
- Modify: `src/m3u.go:233-399`
- Test: `src/m3u_publication_test.go`
- Preserve test: `src/persistence_test.go:96-106`

**Interfaces:**
- Consumes: `publishM3UFile`; existing `buildM3U` parsed/sorted channel state.
- Produces: unchanged `buildM3U([]string{}) == ("", nil)` on successful complete publication and unchanged in-memory filtered string result.

- [ ] **Step 1: Run existing M3U output characterizations before refactor**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestCreateM3UFileReturnsStreamingURLPersistenceError|TestParseGroupCountLabel|TestCompareChannelNumbers)$'
```

Expected: PASS on current `main`. These tests lock existing output and persistence behavior while Task 1's new tests lock publication semantics.

- [ ] **Step 2: Extract existing output loop into local renderer**

Inside `buildM3U`, keep all parsing, group counting, sorting, clamping, URL creation, duplicate detection, and current data selection unchanged. Define `render := func(output io.StringWriter) error` immediately after sorting `entries`. Its first operation calls `output.WriteString(header)` and returns that error. Move the current block beginning with `seenURLInGroup := make(map[string]struct{})` and ending after the channel-entry loop into this closure. Replace each `writer.WriteString` call with `output.WriteString`, remove the `writer != nil` branch, and always write `parameter` followed by `stream + "\n"`. Preserve every existing condition, URL transformation, counter update, and error return byte-for-byte. End the closure with `return nil`.

Use complete mode:

```go
filename := filepath.Join(System.Folder.Data, "threadfin.m3u")
if len(groups) == 0 {
	err = publishM3UFile(filename, render)
	return "", err
}
```

Use filtered mode:

```go
var output strings.Builder
if err = render(&output); err != nil {
	return "", err
}
return output.String(), nil
```

Remove only direct `os.Create`, deferred final-file close, and trailing writer flush. Do not change output bytes.

- [ ] **Step 3: Run M3U and persistence tests; verify GREEN**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestPublishM3UFile|TestBuildM3U|TestCreateM3UFileReturnsStreamingURLPersistenceError|TestParseGroupCountLabel|TestCompareChannelNumbers)$'
```

Expected: PASS.

- [ ] **Step 4: Commit task only when explicitly requested**

```sh
git add src/m3u.go src/m3u_publication_test.go
git commit -m "fix: stream complete M3U through atomic publication"
```

---

### Task 3: Isolate complete versus filtered HTTP response selection

**Files:**
- Modify: `src/webserver.go:3-21`
- Modify: `src/webserver.go:279-377`
- Create: `src/webserver_m3u_test.go`

**Interfaces:**
- Consumes: `buildM3U`, `httpStatusError`, `getFilenameFromPath`, `System.Dev`, and final cache path.
- Produces: `m3uReadFile`, `m3uHandlerOps`, `defaultM3UHandlerOps`, `serveM3U`, and `serveOpenedM3U`.

- [ ] **Step 1: Add handler test setup without parallel execution**

```go
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

func requestThreadfinM3U(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()
	Threadfin(response, request)
	return response
}
```

Add this authentication helper and import `threadfin/src/internal/authentication`:

```go
func enableM3UAuthentication(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	if err := authentication.Init(filepath.Join(root, "config"), 60); err != nil {
		t.Fatal(err)
	}
	userID, err := authentication.CreateNewUser("playlist-user", "playlist-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := authentication.WriteUserData(userID, map[string]interface{}{
		"authentication.m3u": true,
	}); err != nil {
		t.Fatal(err)
	}
	Settings.AuthenticationM3U = true
}
```

- [ ] **Step 2: Write failing black-box handler tests**

Cover:

1. Existing complete cache with no query returns exact bytes.
2. Existing complete cache with unknown query returns same bytes without rebuild.
3. Existing complete cache with valid `username`/`password` returns same bytes.
4. Wrong credentials return 403 without exposing cache bytes.
5. Missing complete cache builds and returns published bytes in same request.
6. Cache path that is a directory returns 500.
7. Missing/unwritable data directory returns non-empty 500.

```go
func TestThreadfinServesCachedCompleteM3UWithUnknownQuery(t *testing.T) {
	filename := setupM3UDeliveryTest(t)
	const cached = "#EXTM3U\ncached playlist\n"
	if err := os.WriteFile(filename, []byte(cached), 0600); err != nil {
		t.Fatal(err)
	}

	response := requestThreadfinM3U(t, "/m3u/threadfin.m3u?client=regression")
	if response.Code != http.StatusOK || response.Body.String() != cached {
		t.Fatalf("response = (%d, %q), want (200, %q)", response.Code, response.Body.String(), cached)
	}
}
```

- [ ] **Step 3: Run handler tests and verify RED**

```sh
go test -count=1 -mod=vendor ./src -run '^TestThreadfin.*M3U'
```

Expected: authenticated/unknown-query/missing-file/error cases fail against query-count routing and empty complete-build result.

- [ ] **Step 4: Add failing dependency-seam tests**

Define seam tests before implementation:

```go
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
```

Add seams for initial non-not-exist open error, successful build followed by open error, filtered build error, non-regular opened file, and opened-file `Stat` error. Assert no successful content is written before each returned error.

- [ ] **Step 5: Run seam tests and verify RED**

```sh
go test -count=1 -mod=vendor ./src -run '^TestServeM3U'
```

Expected: compile failure because helper types/functions do not exist.

- [ ] **Step 6: Implement opened-file and handler operations**

Add imports `errors`, `io`, `io/fs`, and `path/filepath` as needed.

```go
type m3uReadFile interface {
	io.ReadSeeker
	Stat() (os.FileInfo, error)
	Close() error
}

type m3uHandlerOps struct {
	build func([]string) (string, error)
	open  func(string) (m3uReadFile, error)
}

func defaultM3UHandlerOps() m3uHandlerOps {
	return m3uHandlerOps{
		build: buildM3U,
		open: func(filename string) (m3uReadFile, error) {
			return os.Open(filename)
		},
	}
}
```

Implement `serveOpenedM3U`:

```go
func serveOpenedM3U(w http.ResponseWriter, r *http.Request, file m3uReadFile) error {
	info, err := file.Stat()
	if err != nil {
		return errors.Join(fmt.Errorf("inspect complete M3U: %w", err), file.Close())
	}
	if !info.Mode().IsRegular() {
		return errors.Join(fmt.Errorf("complete M3U %q is not a regular file", info.Name()), file.Close())
	}

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
	if err := file.Close(); err != nil {
		ShowError(err, 0)
	}
	return nil
}
```

Implement `serveM3U`:

- For non-empty `groupTitle`, split by comma, build in memory, set existing generated/filtered download and content-type headers, write content, return.
- For empty `groupTitle`, open final cache.
- Only `errors.Is(err, fs.ErrNotExist)` triggers `ops.build(nil)`.
- Set existing generated-path download header before build.
- Reopen generated cache and pass it to `serveOpenedM3U`.
- Existing cache goes directly to `serveOpenedM3U` without changing current cached-path attachment behavior.

- [ ] **Step 7: Integrate helper into `Threadfin` after authentication**

Keep `urlAuth` first, then read `group-title` once:

```go
groupTitle = r.URL.Query().Get("group-title")

systemMutex.Lock()
m3uFilePath := filepath.Join(System.Folder.Data, "threadfin.m3u")
systemMutex.Unlock()

err = serveM3U(w, r, m3uFilePath, groupTitle, defaultM3UHandlerOps())
if err != nil {
	ShowError(err, 0)
	httpStatusError(w, r, http.StatusInternalServerError)
}
return
```

Remove query-count cache selection and direct generic `buildM3U`/content write only from M3U branch. Do not change XMLTV branch.

- [ ] **Step 8: Run handler and seam tests; verify GREEN**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestThreadfin.*M3U|TestServeM3U)'
```

Expected: PASS.

- [ ] **Step 9: Commit task only when explicitly requested**

```sh
git add src/webserver.go src/webserver_m3u_test.go
git commit -m "fix: deliver complete M3U in authenticated requests"
```

---

### Task 4: Verify M3U workstream

**Files:**
- Verify: `src/m3u.go`
- Verify: `src/m3u_publication_test.go`
- Verify: `src/webserver.go`
- Verify: `src/webserver_m3u_test.go`

**Interfaces:**
- Consumes: complete M3U workstream.
- Produces: formatted, tested, vetted, cross-built change without unrelated behavior drift.

- [ ] **Step 1: Format changed Go files**

```sh
gofmt -w src/m3u.go src/m3u_publication_test.go src/webserver.go src/webserver_m3u_test.go
```

- [ ] **Step 2: Run focused regressions**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestPublishM3UFile|TestBuildM3U|TestThreadfin.*M3U|TestServeM3U|TestCreateM3UFileReturnsStreamingURLPersistenceError|TestParseGroupCountLabel|TestCompareChannelNumbers|TestCheckConditionsSeparatorCompatibility)$'
```

Expected: PASS.

- [ ] **Step 3: Run repository tests and vet**

```sh
go test -count=1 -mod=vendor ./...
go vet -mod=vendor ./...
```

Expected: both commands exit 0.

- [ ] **Step 4: Run required Linux builds**

```sh
env GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-s -w" -o /tmp/Threadfin_linux_amd64 .
env GOOS=linux GOARCH=arm64 go build -mod=vendor -ldflags="-s -w" -o /tmp/Threadfin_linux_arm64 .
```

Expected: both builds exit 0.

- [ ] **Step 5: Check scope and diff**

```sh
git diff --check -- src/m3u.go src/m3u_publication_test.go src/webserver.go src/webserver_m3u_test.go
git diff -- src/m3u.go src/m3u_publication_test.go src/webserver.go src/webserver_m3u_test.go
git status --short
```

Confirm authentication policy, M3U bytes, group syntax, XEPG ordering, dependencies, vendor files, assets, and release metadata remain unchanged.

- [ ] **Step 6: Commit workstream only when explicitly requested**

```sh
git add src/m3u.go src/m3u_publication_test.go src/webserver.go src/webserver_m3u_test.go
git commit -m "fix: deliver complete M3U reliably"
```
