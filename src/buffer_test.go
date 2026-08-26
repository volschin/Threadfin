package src

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

func TestTransferSegmentClosesThenRunsCallbackThenWrites(t *testing.T) {
	source := &trackingSegmentReadCloser{Reader: strings.NewReader("segment-data")}
	callbackRan := false
	destination := segmentWriterFunc(func(content []byte) (int, error) {
		if !source.closed {
			t.Fatal("destination write occurred before close")
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
		if !source.closed {
			t.Fatal("callback occurred before close")
		}
		if !bytes.Equal(content, []byte("segment-data")) {
			t.Fatalf("callback content = %q", content)
		}
		callbackRan = true
	})
	if inputErr != nil || writeErr != nil {
		t.Fatalf("transferSegment() = (%v, %v), want (nil, nil)", inputErr, writeErr)
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
