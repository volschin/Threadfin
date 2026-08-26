package src

import (
	"bytes"
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

func TestMapToJSONUsesCompactEncoding(t *testing.T) {
	value := struct {
		Map   map[string]string `json:"map"`
		Slice []string          `json:"slice"`
	}{}
	if got, want := mapToJSON(value), `{"map":null,"slice":null}`; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := mapToJSON(make(chan int)); got != "{}" {
		t.Fatalf("fallback = %q", got)
	}
}

func TestSaveMapToJSONFileRemainsIndented(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := saveMapToJSONFile(path, map[string]interface{}{"enabled": true}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("{\n  \"enabled\": true\n}")
	if !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}
