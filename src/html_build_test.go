package src

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileToBase64ReturnsOpenError(t *testing.T) {
	_, err := fileToBase64(filepath.Join(t.TempDir(), "missing.html"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fileToBase64() error = %v, want os.ErrNotExist", err)
	}
}
