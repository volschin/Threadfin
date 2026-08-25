package up2date

import (
	"archive/zip"
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
