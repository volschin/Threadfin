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

func TestInterfaceToMapReturnsJSONEncodingError(t *testing.T) {
	_, err := interfaceToMap(map[string]interface{}{"unsupported": make(chan int)})
	if err == nil {
		t.Fatal("interfaceToMap() error = nil, want JSON encoding error")
	}
}

func TestJSONToMapReturnsDecodeError(t *testing.T) {
	_, err := jsonToMap("{not-json")
	if err == nil {
		t.Fatal("jsonToMap() error = nil, want JSON decode error")
	}
}
