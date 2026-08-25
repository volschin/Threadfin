package src

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONFileToMapReturnsOpenError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")

	_, err := loadJSONFileToMap(missing)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("loadJSONFileToMap() error = %v, want os.ErrNotExist", err)
	}
}

func TestReadByteFromFileReturnsOpenError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.bin")

	_, err := readByteFromFile(missing)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readByteFromFile() error = %v, want os.ErrNotExist", err)
	}
}
