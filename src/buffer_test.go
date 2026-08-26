package src

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
)

func TestTerminateProcessKillsAndReapsChild(t *testing.T) {
	if os.Getenv("THREADFIN_PROCESS_HELPER") == "1" {
		time.Sleep(time.Minute)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestTerminateProcessKillsAndReapsChild$")
	cmd.Env = append(os.Environ(), "THREADFIN_PROCESS_HELPER=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	if err := terminateProcess(cmd); err != nil {
		t.Fatalf("terminateProcess() error = %v", err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("terminateProcess() did not reap the child process")
	}
}

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

func TestCreateBufferFileReturnsCreateError(t *testing.T) {
	useMemoryBufferVFS(t)

	err := createBufferFile(filepath.Join("missing", "segment.ts"))
	if err == nil {
		t.Fatal("createBufferFile() error = nil, want missing-directory error")
	}
}

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

func TestPrepareSegmentDirectoryBackupStartsAtOneWithoutRetainedSegments(t *testing.T) {
	for _, test := range []struct {
		name            string
		createDirectory bool
	}{
		{name: "missing directory"},
		{name: "empty directory", createDirectory: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			vfs := useMemoryBufferVFS(t)
			folder := "buffer" + string(os.PathSeparator)
			if test.createDirectory {
				if err := vfs.MkdirAll(folder, 0755); err != nil {
					t.Fatal(err)
				}
			}

			start, err := prepareSegmentDirectory(folder, true)
			if err != nil {
				t.Fatalf("prepareSegmentDirectory() error = %v", err)
			}
			if start != 1 {
				t.Fatalf("prepareSegmentDirectory() start = %d, want 1", start)
			}
		})
	}
}

func TestPrepareSegmentDirectoryBackupContinuesAcrossRepeatedTransitions(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	for segment := 1; segment <= 5; segment++ {
		writeBufferTestFile(t, vfs, filepath.Join(folder, strconv.Itoa(segment)+".ts"), "retained")
	}

	start, err := prepareSegmentDirectory(folder, true)
	if err != nil {
		t.Fatalf("prepareSegmentDirectory() error = %v", err)
	}
	if start != 6 {
		t.Fatalf("prepareSegmentDirectory() start = %d, want 6", start)
	}
	writeBufferTestFile(t, vfs, filepath.Join(folder, "6.ts"), "retained")

	start, err = prepareSegmentDirectory(folder, true)
	if err != nil {
		t.Fatalf("prepareSegmentDirectory() error = %v", err)
	}
	if start != 7 {
		t.Fatalf("prepareSegmentDirectory() start = %d, want 7", start)
	}
}

func TestPrepareSegmentDirectoryBackupIgnoresNoncanonicalFiles(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	outOfRangeName := strconv.FormatUint(uint64(^uint(0)>>1)+1, 10) + ".ts"
	ignoredNames := []string{"01.ts", "+2.ts", "-3.ts", "0.ts", "3.0.ts", "4.ts.bak", "notes.txt", " 6.ts", outOfRangeName}
	for _, name := range ignoredNames {
		writeBufferTestFile(t, vfs, filepath.Join(folder, name), "retained")
	}
	ignoredDirectory := filepath.Join(folder, "999.ts")
	if err := vfs.Mkdir(ignoredDirectory, 0755); err != nil {
		t.Fatal(err)
	}
	writeBufferTestFile(t, vfs, filepath.Join(folder, "5.ts"), "retained")

	start, err := prepareSegmentDirectory(folder, true)
	if err != nil {
		t.Fatalf("prepareSegmentDirectory() error = %v", err)
	}
	if start != 6 {
		t.Fatalf("prepareSegmentDirectory() start = %d, want 6", start)
	}
	for _, name := range ignoredNames {
		if _, err := vfs.Stat(filepath.Join(folder, name)); err != nil {
			t.Fatalf("ignored entry %q was not retained: %v", name, err)
		}
	}
	if info, err := vfs.Stat(ignoredDirectory); err != nil || !info.IsDir() {
		t.Fatalf("ignored directory = (%v, %v), want retained directory", info, err)
	}
}

func TestPrepareSegmentDirectoryBackupReturnsReadDirError(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	retainedPath := filepath.Join(folder, "7.ts")
	writeBufferTestFile(t, vfs, retainedPath, "retained")
	readErr := errors.New("read directory failed")
	bufferVFS = readDirErrorVFS{VFS: vfs, err: readErr}

	start, err := prepareSegmentDirectory(folder, true)
	if start != 0 || !errors.Is(err, readErr) {
		t.Fatalf("prepareSegmentDirectory() = (%d, %v), want (0, %v)", start, err, readErr)
	}
	content, err := vfs.ReadFile(retainedPath)
	if err != nil || string(content) != "retained" {
		t.Fatalf("retained segment after ReadDir error = (%q, %v), want (retained, nil)", content, err)
	}
}

func TestPrepareSegmentDirectoryBackupReturnsOverflowForMaxIntSegment(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	maxInt := int(^uint(0) >> 1)
	retainedPath := filepath.Join(folder, strconv.Itoa(maxInt)+".ts")
	writeBufferTestFile(t, vfs, retainedPath, "retained")

	start, err := prepareSegmentDirectory(folder, true)
	if start != maxInt || !errors.Is(err, errSegmentNumberOverflow) {
		t.Fatalf("prepareSegmentDirectory() = (%d, %v), want (%d, %v)", start, err, maxInt, errSegmentNumberOverflow)
	}
	content, err := vfs.ReadFile(retainedPath)
	if err != nil || string(content) != "retained" {
		t.Fatalf("retained max segment after overflow = (%q, %v), want (retained, nil)", content, err)
	}
}

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

func TestGetBufTmpFilesReturnsFirstCompletedSegmentWithCurrentSuccessor(t *testing.T) {
	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	writeBufferTestFile(t, vfs, filepath.Join(folder, "1.ts"), "complete")
	writeBufferTestFile(t, vfs, filepath.Join(folder, "2.ts"), "in progress")

	stream := ThisStream{Folder: folder}
	if got := getBufTmpFiles(&stream); !slices.Equal(got, []string{"1.ts"}) {
		t.Fatalf("getBufTmpFiles() = %v, want [1.ts]", got)
	}
	if !slices.Equal(stream.OldSegments, []string{"1.ts"}) {
		t.Fatalf("OldSegments = %v, want [1.ts]", stream.OldSegments)
	}
}

func TestGetBufTmpFilesPreservesAdjacentLargeSegmentBasenames(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("requires 64-bit native ints")
	}

	vfs := useMemoryBufferVFS(t)
	folder := "buffer" + string(os.PathSeparator)
	if err := vfs.MkdirAll(folder, 0755); err != nil {
		t.Fatal(err)
	}
	completed := []string{"9007199254740992.ts", "9007199254740993.ts"}
	for _, name := range append(slices.Clone(completed), "9007199254740994.ts") {
		writeBufferTestFile(t, vfs, filepath.Join(folder, name), "")
	}

	stream := ThisStream{Folder: folder}
	if got := getBufTmpFiles(&stream); !slices.Equal(got, completed) {
		t.Fatalf("getBufTmpFiles() = %v, want %v", got, completed)
	}
	if !slices.Equal(stream.OldSegments, completed) {
		t.Fatalf("OldSegments = %v, want %v", stream.OldSegments, completed)
	}
}

type trackingSegmentReadCloser struct {
	io.Reader
	closed   bool
	closeErr error
}

func (reader *trackingSegmentReadCloser) Close() error {
	reader.closed = true
	return reader.closeErr
}

type segmentWriterFunc func([]byte) (int, error)

func (write segmentWriterFunc) Write(content []byte) (int, error) {
	return write(content)
}

type segmentErrorReader struct {
	err error
}

func (reader segmentErrorReader) Read([]byte) (int, error) {
	return 0, reader.err
}

func TestTransferSegmentRunsCallbackThenWritesThenCloses(t *testing.T) {
	source := &trackingSegmentReadCloser{Reader: strings.NewReader("segment-data")}
	callbackRan := false
	destination := segmentWriterFunc(func(content []byte) (int, error) {
		if source.closed {
			t.Fatal("destination write occurred after close")
		}
		if !callbackRan {
			t.Fatal("destination write occurred before callback")
		}
		if !bytes.Equal(content, []byte("segment-data")) {
			t.Fatalf("written content = %q", content)
		}
		return len(content), nil
	})

	inputErr, writeErr := transferSegment(destination, source, func(content []byte) {
		if source.closed {
			t.Fatal("callback occurred after close")
		}
		if !bytes.Equal(content, []byte("segment-data")) {
			t.Fatalf("callback content = %q", content)
		}
		callbackRan = true
	})
	if inputErr != nil || writeErr != nil {
		t.Fatalf("transferSegment() = (%v, %v), want (nil, nil)", inputErr, writeErr)
	}
	if !source.closed {
		t.Fatal("source was not closed")
	}
}

func TestTransferSegmentJoinsInputErrorsAndSkipsCallbackAndWrite(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	source := &trackingSegmentReadCloser{
		Reader:   segmentErrorReader{err: readErr},
		closeErr: closeErr,
	}
	callbackRan := false
	writeRan := false
	destination := segmentWriterFunc(func(content []byte) (int, error) {
		writeRan = true
		return len(content), nil
	})

	inputErr, writeErr := transferSegment(destination, source, func([]byte) {
		callbackRan = true
	})
	if !errors.Is(inputErr, readErr) || !errors.Is(inputErr, closeErr) {
		t.Fatalf("input error = %v, want joined read and close errors", inputErr)
	}
	if writeErr != nil {
		t.Fatalf("write error = %v, want nil", writeErr)
	}
	if !source.closed || callbackRan || writeRan {
		t.Fatalf("closed=%t callback=%t write=%t, want true false false", source.closed, callbackRan, writeRan)
	}
}

func TestTransferSegmentReturnsWriterErrorSeparately(t *testing.T) {
	writeErr := errors.New("write failed")
	source := &trackingSegmentReadCloser{Reader: strings.NewReader("segment")}
	inputErr, gotWriteErr := transferSegment(
		segmentWriterFunc(func([]byte) (int, error) { return 0, writeErr }),
		source,
		nil,
	)
	if inputErr != nil || !errors.Is(gotWriteErr, writeErr) {
		t.Fatalf("transferSegment() = (%v, %v), want (nil, %v)", inputErr, gotWriteErr, writeErr)
	}
}

func TestTransferSegmentPreservesNilErrorOnShortWrite(t *testing.T) {
	source := &trackingSegmentReadCloser{Reader: strings.NewReader("segment")}
	inputErr, writeErr := transferSegment(
		segmentWriterFunc(func([]byte) (int, error) { return 1, nil }),
		source,
		nil,
	)
	if inputErr != nil || writeErr != nil {
		t.Fatalf("transferSegment() = (%v, %v), want (nil, nil)", inputErr, writeErr)
	}
}

func TestSwitchBandwidthPreservesLegacyVariantSelection(t *testing.T) {
	stream := &ThisStream{
		NetworkBandwidth: 2500,
		DynamicStream: map[int]DynamicStream{
			3000: {Bandwidth: 3000, URL: "high"},
			1000: {Bandwidth: 1000, URL: "low"},
			2000: {Bandwidth: 2000, URL: "middle"},
		},
	}

	if err := switchBandwidth(stream); err != nil {
		t.Fatalf("switchBandwidth() error = %v", err)
	}
	if len(stream.Segment) != 1 {
		t.Fatalf("switchBandwidth() appended %d segments, want 1", len(stream.Segment))
	}
	if got := stream.Segment[0].URL; got != "low" {
		t.Fatalf("selected URL = %q, want legacy fallback %q", got, "low")
	}
	if got := stream.Segment[0].StreamInf.Bandwidth; got != 3000 {
		t.Fatalf("segment bandwidth = %d, want %d", got, 3000)
	}
}

func TestSwitchBandwidthUsesLowestVariantWhenBandwidthUnknown(t *testing.T) {
	stream := &ThisStream{
		DynamicStream: map[int]DynamicStream{
			3000: {Bandwidth: 3000, URL: "high"},
			1000: {Bandwidth: 1000, URL: "low"},
		},
	}

	if err := switchBandwidth(stream); err != nil {
		t.Fatalf("switchBandwidth() error = %v", err)
	}
	if len(stream.Segment) != 1 {
		t.Fatalf("switchBandwidth() appended %d segments, want 1", len(stream.Segment))
	}
	if got := stream.Segment[0].URL; got != "low" {
		t.Fatalf("selected URL = %q, want %q", got, "low")
	}
}

func TestSwitchBandwidthRejectsEmptyVariantMap(t *testing.T) {
	stream := &ThisStream{DynamicStream: map[int]DynamicStream{}}

	if err := switchBandwidth(stream); err == nil {
		t.Fatal("switchBandwidth() succeeded with no streaming variants")
	}
}

var errSegmentRead = errors.New("segment read failed")
var errSegmentWrite = errors.New("segment write failed")
var errSegmentClose = errors.New("segment close failed")

type segmentBody struct {
	io.Reader
	closeErr error
	closes   int
}

func (body *segmentBody) Close() error {
	body.closes++
	return body.closeErr
}

type segmentMidReadFailure struct {
	remaining int
}

func (reader *segmentMidReadFailure) Read(buffer []byte) (int, error) {
	if reader.remaining == 0 {
		return 0, errSegmentRead
	}
	n := min(len(buffer), reader.remaining)
	for index := range n {
		buffer[index] = 0x47
	}
	reader.remaining -= n
	return n, nil
}

type segmentWriteFailure struct {
	limit   int
	written int
}

func (writer *segmentWriteFailure) Write(buffer []byte) (int, error) {
	remaining := writer.limit - writer.written
	if remaining <= 0 {
		return 0, errSegmentWrite
	}
	if len(buffer) > remaining {
		writer.written += remaining
		return remaining, errSegmentWrite
	}
	writer.written += len(buffer)
	return len(buffer), nil
}

func TestTransferSegmentStreamsExactBytesAfterPrefix(t *testing.T) {
	for _, size := range []int{0, 100, 512, 1024} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			payload := bytes.Repeat([]byte{0x47}, size)
			source := &segmentBody{Reader: bytes.NewReader(payload)}
			var destination bytes.Buffer
			var prefix []byte
			inputErr, writeErr := transferSegment(&destination, source, func(value []byte) {
				prefix = append([]byte(nil), value...)
			})
			if inputErr != nil || writeErr != nil {
				t.Fatalf("errors=(%v,%v)", inputErr, writeErr)
			}
			if !bytes.Equal(destination.Bytes(), payload) {
				t.Fatal("segment bytes changed")
			}
			if !bytes.Equal(prefix, payload[:min(len(payload), 512)]) {
				t.Fatal("prefix changed")
			}
			if source.closes != 1 {
				t.Fatalf("Close calls=%d", source.closes)
			}
		})
	}
}

func TestTransferSegmentExposesPartialBytesAfterPrefixReadFailure(t *testing.T) {
	source := &segmentBody{Reader: &segmentMidReadFailure{remaining: 1024}, closeErr: errSegmentClose}
	var destination bytes.Buffer
	inputErr, writeErr := transferSegment(&destination, source, func([]byte) {})
	if !errors.Is(inputErr, errSegmentRead) || !errors.Is(inputErr, errSegmentClose) || writeErr != nil {
		t.Fatalf("errors=(%v,%v)", inputErr, writeErr)
	}
	if destination.Len() != 1024 {
		t.Fatalf("partial bytes=%d", destination.Len())
	}
}

func TestTransferSegmentReturnsWriteAndCloseErrors(t *testing.T) {
	source := &segmentBody{Reader: bytes.NewReader(bytes.Repeat([]byte{0x47}, 2048)), closeErr: errSegmentClose}
	destination := &segmentWriteFailure{limit: 700}
	inputErr, writeErr := transferSegment(destination, source, func([]byte) {})
	if !errors.Is(inputErr, errSegmentClose) || !errors.Is(writeErr, errSegmentWrite) {
		t.Fatalf("errors=(%v,%v)", inputErr, writeErr)
	}
	if destination.written != 700 {
		t.Fatalf("written=%d", destination.written)
	}
}

func TestTransferSegmentDoesNotWriteWhenPrefixReadFails(t *testing.T) {
	source := &segmentBody{Reader: &segmentMidReadFailure{remaining: 100}}
	var destination bytes.Buffer
	called := false
	inputErr, writeErr := transferSegment(&destination, source, func([]byte) { called = true })
	if !errors.Is(inputErr, errSegmentRead) || writeErr != nil {
		t.Fatalf("errors=(%v,%v)", inputErr, writeErr)
	}
	if destination.Len() != 0 || called {
		t.Fatal("prefix failure committed output")
	}
}
