package src

import (
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

var errArchiveWrite = errors.New("archive write failed")

type archiveErrorWriter struct{}

func (archiveErrorWriter) Write([]byte) (int, error) {
	return 0, errArchiveWrite
}

func TestZipFilesReturnsMissingSourceError(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "backup.zip")
	missing := filepath.Join(root, "missing")

	if err := zipFiles([]string{missing}, target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("zipFiles() error = %v, want os.ErrNotExist", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("failed zipFiles() left partial archive: %v", err)
	}
}

func TestWriteZIPReturnsFinalizationError(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	if err := os.WriteFile(source, []byte("backup data"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := writeZIP([]string{source}, archiveErrorWriter{}); !errors.Is(err, errArchiveWrite) {
		t.Fatalf("writeZIP() error = %v, want %v", err, errArchiveWrite)
	}
}

func TestNewHTTPServerConfiguresSlowClientTimeouts(t *testing.T) {
	server := newHTTPServer("127.0.0.1:0", http.NewServeMux())

	if server.ReadHeaderTimeout <= 0 {
		t.Error("ReadHeaderTimeout must reject clients that never finish headers")
	}
	if server.IdleTimeout <= 0 {
		t.Error("IdleTimeout must close abandoned keep-alive connections")
	}
}

func TestServeHTTPServerSignalsReadinessOnlyAfterListenerAcquisition(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := probe.Addr().String()
	if err := probe.Close(); err != nil {
		t.Fatal(err)
	}

	server := newHTTPServer(address, http.NewServeMux())
	readyErr := errors.New("stop after readiness assertion")
	readyCalled := false
	err = serveHTTPServer(server, func() error {
		readyCalled = true
		second, listenErr := net.Listen("tcp", address)
		if listenErr == nil {
			_ = second.Close()
			t.Fatal("readiness was signaled before the HTTP listener was acquired")
		}
		return readyErr
	})
	if !errors.Is(err, readyErr) {
		t.Fatalf("serve error = %v, want readiness error", err)
	}
	if !readyCalled {
		t.Fatal("readiness was not signaled after listener acquisition")
	}
}

func TestServeHTTPServerDoesNotSignalReadinessWhenListenerAcquisitionFails(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()

	server := newHTTPServer(occupied.Addr().String(), http.NewServeMux())
	readyCalled := false
	if err := serveHTTPServer(server, func() error {
		readyCalled = true
		return nil
	}); err == nil {
		t.Fatal("server acquired an already occupied listener")
	}
	if readyCalled {
		t.Fatal("readiness was signaled without listener ownership")
	}
}
