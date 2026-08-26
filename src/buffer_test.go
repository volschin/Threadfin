package src

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

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

func TestCreateBufferFileReturnsCreateError(t *testing.T) {
	previousVFS := bufferVFS
	bufferVFS = memfs.New()
	t.Cleanup(func() { bufferVFS = previousVFS })

	err := createBufferFile(filepath.Join("missing", "segment.ts"))
	if err == nil {
		t.Fatal("createBufferFile() error = nil, want missing-directory error")
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
