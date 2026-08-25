package src

import (
	"archive/zip"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZIPRejectsParentTraversal(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "malicious.zip")
	target := filepath.Join(root, "target")
	escapedPath := filepath.Join(root, "escaped.txt")

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	archive := zip.NewWriter(archiveFile)
	entry, err := archive.Create("../escaped.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("escaped")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := extractZIP(archivePath, target); err == nil {
		t.Fatal("extractZIP accepted an entry outside the target directory")
	}
	if _, err := os.Stat(escapedPath); !os.IsNotExist(err) {
		t.Fatalf("archive entry escaped the target directory: %v", err)
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
