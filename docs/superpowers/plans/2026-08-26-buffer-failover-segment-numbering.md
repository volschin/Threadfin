# Buffer Failover Segment Numbering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Preserve numeric segment progression across backup transitions while retained buffer files exist, without changing primary reset behavior or persisting `OldSegments`.

**Architecture:** Add small unexported helpers around the existing virtual filesystem boundary to prepare primary/backup directories, parse canonical numeric segment names, and guard counter increments. `thirdPartyBuffer` consumes the selected `startSegment`, while existing delivery, process, client, and backup-selection flows remain intact.

**Tech Stack:** Go 1.27, `github.com/avfs/avfs` virtual filesystem, `github.com/avfs/avfs/vfs/memfs`, standard `testing` package

**Spec:** `docs/superpowers/specs/2026-08-26-stream-reliability-fixes-design.md`

## Global Constraints

- Work against current `main`; preserve `slices.Index` duplicate detection and streamed `transferSegment` behavior.
- Directory contents are the only numbering source. Do not persist `OldSegments` into shared stream state and do not add a lifetime counter.
- A primary invocation still clears its directory and starts at `1.ts`.
- A backup invocation never clears its directory and starts above its highest retained canonical numeric `.ts` regular file.
- Guard startup and every later increment against `int` overflow; route failures through existing stream fallback behavior.
- Preserve backup ordering, 20-second startup timeout, process termination, client notification, buffer thresholds, configuration, and persisted formats.
- Do not change dependencies, generated assets, versions, release metadata, or XEPG behavior.
- Run all tests and builds with checked-in vendor tree.

## File Structure

- Modify `src/buffer.go`: add segment-number helpers and integrate them into `thirdPartyBuffer`.
- Modify `src/buffer_test.go`: add VFS fixtures, directory-policy tests, overflow tests, invocation-relative first-segment tests, and delivery characterization.
- Read-only context: `src/config.go` owns `bufferVFS`; `src/toolchain.go` owns `checkVFSFolder`; `src/struct-buffer.go` defines `ThisStream`.

---

### Task 1: Isolate directory preparation and retained-file discovery

**Files:**
- Modify: `src/buffer.go:780-895`
- Test: `src/buffer_test.go:3-47`

**Interfaces:**
- Consumes: global `bufferVFS avfs.VFS`; `checkVFSFolder(path string, vfs avfs.VFS) error`; `getPlatformPath(string) string`.
- Produces: `incrementSegmentNumber(segment int) (int, error)`, `nextBackupSegment(folder string) (int, error)`, `prepareSegmentDirectory(folder string, useBackup bool) (int, error)`, and sentinel `errSegmentNumberOverflow`.

- [ ] **Step 1: Add reusable memory-VFS test fixtures**

Add `io/fs`, `path/filepath`, `strconv`, and `github.com/avfs/avfs` imports as needed. Keep tests serial because `bufferVFS` is global.

```go
type readDirErrorVFS struct {
	avfs.VFS
	err error
}

func (vfs readDirErrorVFS) ReadDir(string) ([]fs.DirEntry, error) {
	return nil, vfs.err
}

func useMemoryBufferVFS(t *testing.T) avfs.VFS {
	t.Helper()

	previousVFS := bufferVFS
	vfs := memfs.New()
	bufferVFS = vfs
	t.Cleanup(func() {
		bufferVFS = previousVFS
	})

	return vfs
}

func writeBufferTestFile(t *testing.T, vfs avfs.VFS, path, content string) {
	t.Helper()

	if err := vfs.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
```

Update `TestCreateBufferFileReturnsCreateError` to use `useMemoryBufferVFS(t)` instead of duplicating global restoration.

- [ ] **Step 2: Write failing primary and backup directory-policy tests**

```go
func TestPrepareSegmentDirectoryPrimaryClearsFilesAndStartsAtOne(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	writeBufferTestFile(t, vfs, filepath.Join(folder, "37.ts"), "stale")
	writeBufferTestFile(t, vfs, filepath.Join(folder, "notes.txt"), "stale")

	start, err := prepareSegmentDirectory(folder, false)
	if err != nil {
		t.Fatalf("prepareSegmentDirectory() error = %v", err)
	}
	if start != 1 {
		t.Fatalf("prepareSegmentDirectory() start = %d, want 1", start)
	}
	entries, err := vfs.ReadDir(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("primary directory contains %d entries after reset, want 0", len(entries))
	}
}

func TestPrepareSegmentDirectoryBackupStartsAfterHighestRetainedFile(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	for _, segment := range []int{4, 7, 11} {
		writeBufferTestFile(t, vfs, filepath.Join(folder, strconv.Itoa(segment)+".ts"), "retained")
	}

	start, err := prepareSegmentDirectory(folder, true)
	if err != nil {
		t.Fatalf("prepareSegmentDirectory() error = %v", err)
	}
	if start != 12 {
		t.Fatalf("prepareSegmentDirectory() start = %d, want 12", start)
	}
	content, err := vfs.ReadFile(filepath.Join(folder, "11.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "retained" {
		t.Fatalf("retained segment content = %q, want %q", content, "retained")
	}
}
```

Add a table-driven missing/empty-directory test expecting `1`, plus a repeated-transition test that writes `1.ts` through `5.ts`, expects `6`, writes `6.ts`, then expects `7`.

- [ ] **Step 3: Run focused tests and verify RED**

Run:

```sh
go test -count=1 -mod=vendor ./src -run '^TestPrepareSegmentDirectory'
```

Expected: build failure because `prepareSegmentDirectory` is undefined.

- [ ] **Step 4: Implement canonical numbering helpers near `thirdPartyBuffer`**

```go
var errSegmentNumberOverflow = errors.New("segment number overflow")

func incrementSegmentNumber(segment int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	if segment == maxInt {
		return segment, errSegmentNumberOverflow
	}
	return segment + 1, nil
}

func nextBackupSegment(folder string) (int, error) {
	entries, err := bufferVFS.ReadDir(getPlatformPath(folder))
	if err != nil {
		return 0, err
	}

	highest := 0
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		numberText, ok := strings.CutSuffix(entry.Name(), ".ts")
		if !ok || numberText == "" {
			continue
		}
		number, err := strconv.Atoi(numberText)
		if err != nil || number < 1 || strconv.Itoa(number) != numberText {
			continue
		}
		if number > highest {
			highest = number
		}
	}
	if highest == 0 {
		return 1, nil
	}
	return incrementSegmentNumber(highest)
}

func prepareSegmentDirectory(folder string, useBackup bool) (int, error) {
	if !useBackup {
		if err := bufferVFS.RemoveAll(getPlatformPath(folder)); err != nil {
			ShowError(err, 4005)
		}
	}
	if err := checkVFSFolder(folder, bufferVFS); err != nil {
		return 0, err
	}
	if !useBackup {
		return 1, nil
	}
	return nextBackupSegment(folder)
}
```

Canonical validation through `strconv.Itoa(number) == numberText` rejects signs, leading zeroes, whitespace, decimals, zero, negatives, overflow, and names not produced by `%d.ts`.

- [ ] **Step 5: Integrate directory preparation into `thirdPartyBuffer`**

Remove unconditional `tmpSegment := 1`, `RemoveAll`, and directory recreation. After `addErrorToStream` is declared, use:

```go
tmpSegment, err := prepareSegmentDirectory(tmpFolder, useBackup)
if err != nil {
	ShowError(err, 0)
	killClientConnection(streamID, playlistID, false)
	addErrorToStream(err)
	return
}
startSegment := tmpSegment
```

Keep primary `RemoveAll` logging behavior. Backup `ReadDir` and startup-overflow errors must enter existing client/fallback path without deleting retained files.

- [ ] **Step 6: Run focused tests and verify GREEN**

Run:

```sh
go test -count=1 -mod=vendor ./src -run '^TestPrepareSegmentDirectory'
```

Expected: PASS.

- [ ] **Step 7: Commit task only when explicitly requested**

```sh
git add src/buffer.go src/buffer_test.go
git commit -m "fix: preserve segment numbering across buffer failover"
```

---

### Task 2: Reject malformed entries and startup overflow

**Files:**
- Modify: `src/buffer.go` helper region from Task 1
- Test: `src/buffer_test.go` after directory-policy tests

**Interfaces:**
- Consumes: Task 1 helpers and `readDirErrorVFS`.
- Produces: tested strict parser behavior and deterministic unreadable-directory/startup-overflow errors.

- [ ] **Step 1: Write failing invalid-entry test**

Create one valid `1.ts`, then add `0009.ts`, `+8.ts`, `-7.ts`, `0.ts`, `6.5.ts`, `7.ts.tmp`, `notes.txt`, an overflowing decimal plus `.ts`, and a directory named `99.ts`. Assert `prepareSegmentDirectory(folder, true)` returns `2`.

```go
maxInt := int(^uint(0) >> 1)
overflowingName := strconv.FormatUint(uint64(maxInt)+1, 10) + ".ts"
```

- [ ] **Step 2: Write failing unreadable-directory and startup-overflow tests**

```go
func TestPrepareSegmentDirectoryBackupReturnsReadDirError(t *testing.T) {
	base := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := base.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	readErr := errors.New("read directory failed")
	bufferVFS = readDirErrorVFS{VFS: base, err: readErr}

	start, err := prepareSegmentDirectory(folder, true)
	if !errors.Is(err, readErr) || start != 0 {
		t.Fatalf("prepareSegmentDirectory() = (%d, %v), want (0, %v)", start, err, readErr)
	}
}

func TestPrepareSegmentDirectoryBackupRejectsStartupOverflow(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	maxInt := int(^uint(0) >> 1)
	writeBufferTestFile(t, vfs, filepath.Join(folder, strconv.Itoa(maxInt)+".ts"), "")

	start, err := prepareSegmentDirectory(folder, true)
	if !errors.Is(err, errSegmentNumberOverflow) || start != maxInt {
		t.Fatalf("prepareSegmentDirectory() = (%d, %v), want (%d, overflow)", start, err, maxInt)
	}
}
```

- [ ] **Step 3: Run edge tests and verify RED or existing helper gaps**

Run:

```sh
go test -count=1 -mod=vendor ./src -run '^TestPrepareSegmentDirectoryBackup(Ignores|Returns|Rejects)'
```

Expected before strict parsing/error propagation is complete: at least one assertion fails.

- [ ] **Step 4: Complete strict parser and error propagation**

Use exact canonical validation shown in Task 1. Do not use `ParseFloat`, suffix replacement, `OldSegments`, or directory deletion in backup discovery. Preserve returned `maxInt` with `errSegmentNumberOverflow` so tests can distinguish exhaustion from absence.

- [ ] **Step 5: Run edge tests and verify GREEN**

Run same command. Expected: PASS.

- [ ] **Step 6: Commit task only when explicitly requested**

```sh
git add src/buffer.go src/buffer_test.go
git commit -m "test: cover failover segment discovery errors"
```

---

### Task 3: Make first-segment checks invocation-relative and guard runtime increments

**Files:**
- Modify: `src/buffer.go:1060-1161`
- Test: `src/buffer_test.go` after segment-discovery tests

**Interfaces:**
- Consumes: `incrementSegmentNumber` and `startSegment` from Task 1.
- Produces: `isFirstSegment(segment, startSegment int) bool`; overflow-safe runtime progression.

- [ ] **Step 1: Write failing helper tests**

```go
func TestIncrementSegmentNumberGuardsOverflow(t *testing.T) {
	next, err := incrementSegmentNumber(6)
	if err != nil || next != 7 {
		t.Fatalf("incrementSegmentNumber(6) = (%d, %v), want (7, nil)", next, err)
	}

	maxInt := int(^uint(0) >> 1)
	next, err = incrementSegmentNumber(maxInt)
	if !errors.Is(err, errSegmentNumberOverflow) || next != maxInt {
		t.Fatalf("incrementSegmentNumber(maxInt) = (%d, %v), want (%d, overflow)", next, err, maxInt)
	}
}

func TestIsFirstSegmentSupportsBackupStart(t *testing.T) {
	if !isFirstSegment(6, 6) {
		t.Fatal("backup start segment was not recognized as first segment")
	}
	if isFirstSegment(7, 6) {
		t.Fatal("later backup segment was recognized as first segment")
	}
}
```

- [ ] **Step 2: Run helper tests and verify RED**

Run:

```sh
go test -count=1 -mod=vendor ./src -run '^(TestIncrementSegmentNumber|TestIsFirstSegment)'
```

Expected: build failure because `isFirstSegment` is undefined.

- [ ] **Step 3: Add first-segment helper and replace literal checks**

```go
func isFirstSegment(segment, startSegment int) bool {
	return segment == startSegment
}
```

Replace both existing literal checks with:

```go
if timeout >= 20 && isFirstSegment(tmpSegment, startSegment) {
```

and:

```go
if isFirstSegment(tmpSegment, startSegment) && !stream.Status {
```

- [ ] **Step 4: Replace runtime `tmpSegment++` with guarded progression**

```go
tmpSegment, err = incrementSegmentNumber(tmpSegment)
if err != nil {
	if processErr := terminateProcess(cmd); processErr != nil {
		ShowError(processErr, 0)
	}
	ShowError(err, 0)
	killClientConnection(streamID, playlistID, false)
	addErrorToStream(err)
	return
}
```

Keep this after current file completion/close and before formatting next filename. Do not alter backup order or recursive `addErrorToStream` behavior.

- [ ] **Step 5: Run helper tests and verify GREEN**

Run same focused command. Expected: PASS.

- [ ] **Step 6: Add delivery characterization test**

```go
func TestGetBufTmpFilesReturnsCompletedNonOverlappingBackupSegment(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	for segment := 1; segment <= 7; segment++ {
		writeBufferTestFile(t, vfs, filepath.Join(folder, strconv.Itoa(segment)+".ts"), "")
	}

	stream := ThisStream{
		Folder:      folder,
		OldSegments: []string{"1.ts", "2.ts", "3.ts", "4.ts", "5.ts"},
	}
	files := getBufTmpFiles(&stream)
	if len(files) != 1 || files[0] != "6.ts" {
		t.Fatalf("getBufTmpFiles() = %v, want [6.ts]", files)
	}
	if stream.OldSegments[len(stream.OldSegments)-1] != "6.ts" {
		t.Fatalf("OldSegments = %v, want newly delivered 6.ts appended", stream.OldSegments)
	}
}
```

This is a characterization test: it preserves local `OldSegments`, sorted delivery, and withholding newest `7.ts`; it must not add shared-state persistence.

- [ ] **Step 7: Run focused buffer tests**

```sh
go test -count=1 -mod=vendor ./src -run '^(TestCreateBufferFile|TestPrepareSegmentDirectory|TestIncrementSegmentNumber|TestIsFirstSegment|TestGetBufTmpFiles|TestTransferSegment|TestSwitchBandwidth)'
```

Expected: PASS.

- [ ] **Step 8: Commit task only when explicitly requested**

```sh
git add src/buffer.go src/buffer_test.go
git commit -m "fix: guard failover segment progression"
```

---

### Task 4: Verify buffer workstream

**Files:**
- Verify: `src/buffer.go`
- Verify: `src/buffer_test.go`

**Interfaces:**
- Consumes: complete buffer workstream.
- Produces: formatted, tested, vetted, cross-built change with no unrelated diff.

- [ ] **Step 1: Format changed Go files**

```sh
gofmt -w src/buffer.go src/buffer_test.go
```

- [ ] **Step 2: Run focused and full tests**

```sh
go test -count=1 -mod=vendor ./src
go test -count=1 -mod=vendor ./...
```

Expected: PASS.

- [ ] **Step 3: Run vet and required Linux builds**

```sh
go vet -mod=vendor ./...
GOOS=linux GOARCH=amd64 go build -mod=vendor ./...
GOOS=linux GOARCH=arm64 go build -mod=vendor ./...
```

Expected: all commands exit 0.

- [ ] **Step 4: Check diff boundaries**

```sh
git diff --check
git diff -- src/buffer.go src/buffer_test.go
```

Expected: no whitespace errors; production/test diff only implements approved buffer workstream and preserves `slices.Index` plus streamed transfer.

- [ ] **Step 5: Commit workstream only when explicitly requested**

```sh
git add src/buffer.go src/buffer_test.go
git commit -m "fix: preserve segment numbering across buffer failover"
```
