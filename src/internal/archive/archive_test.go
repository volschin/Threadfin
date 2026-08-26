package archive

import (
	"archive/zip"
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testEntry struct {
	name string
	body []byte
	mode fs.FileMode
}

func writeTestZIP(t *testing.T, archivePath string, entries ...testEntry) {
	t.Helper()
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	archiveWriter := zip.NewWriter(archiveFile)
	for _, item := range entries {
		header := &zip.FileHeader{Name: item.name, Method: zip.Store}
		header.SetMode(item.mode)
		entryWriter, err := archiveWriter.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entryWriter.Write(item.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := errors.Join(archiveWriter.Close(), archiveFile.Close()); err != nil {
		t.Fatal(err)
	}
}

func corruptFirstEntryData(t *testing.T, archivePath string) {
	t.Helper()
	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 30 || string(data[:4]) != "PK\x03\x04" {
		t.Fatal("archive does not begin with a local ZIP header")
	}
	nameLength := int(binary.LittleEndian.Uint16(data[26:28]))
	extraLength := int(binary.LittleEndian.Uint16(data[28:30]))
	dataOffset := 30 + nameLength + extraLength
	if dataOffset >= len(data) {
		t.Fatal("archive has no entry data to corrupt")
	}
	data[dataOffset] ^= 0xff
	if err := os.WriteFile(archivePath, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZIPCreatesNestedFilesAndOmittedParents(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "nested.zip")
	target := filepath.Join(root, "target")
	writeTestZIP(t, archivePath,
		testEntry{name: "explicit/", mode: fs.ModeDir | 0750},
		testEntry{name: "explicit/file.txt", body: []byte("explicit"), mode: 0640},
		testEntry{name: "implicit/deeper/file.txt", body: []byte("implicit"), mode: 0600},
	)

	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"explicit/file.txt":        "explicit",
		"implicit/deeper/file.txt": "implicit",
	} {
		got, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestExtractZIPOverwritesExistingFile(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "overwrite.zip")
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(target, "threadfin")
	if err := os.WriteFile(existing, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	writeTestZIP(t, archivePath, testEntry{name: "threadfin", body: []byte("new"), mode: 0755})

	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("overwritten file = %q, want %q", got, "new")
	}
}

func TestExtractZIPRejectsTraversalAndSlashAbsolutePaths(t *testing.T) {
	testCases := []struct {
		name       string
		entry      string
		wantReason string
	}{
		{name: "parent", entry: "../escaped.txt", wantReason: "traverses outside target directory"},
		{name: "cleaned parent", entry: "safe/../../escaped.txt", wantReason: "traverses outside target directory"},
		{name: "slash absolute", entry: "/absolute.txt", wantReason: "is absolute"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			archivePath := filepath.Join(root, "malicious.zip")
			target := filepath.Join(root, "target")
			writeTestZIP(t, archivePath, testEntry{name: testCase.entry, body: []byte("escaped"), mode: 0644})

			err := ExtractZIP(archivePath, target)
			if err == nil {
				t.Fatalf("ExtractZIP accepted %q", testCase.entry)
			}
			if !strings.Contains(err.Error(), testCase.entry) || !strings.Contains(err.Error(), testCase.wantReason) {
				t.Fatalf("error %q does not identify %q and reason %q", err, testCase.entry, testCase.wantReason)
			}
			entries, readErr := os.ReadDir(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("rejected archive created entries: %v", entries)
			}
		})
	}
}

func TestExtractZIPDoesNotCreateTargetWhenArchiveCannotOpen(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	err := ExtractZIP(filepath.Join(root, "missing.zip"), target)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ExtractZIP error = %v, want fs.ErrNotExist", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("unreadable archive created target: %v", err)
	}
}

func TestExtractZIPRejectsTargetThatIsAFile(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "archive.zip")
	target := filepath.Join(root, "target")
	writeTestZIP(t, archivePath, testEntry{name: "file.txt", body: []byte("data"), mode: 0644})
	if err := os.WriteFile(target, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExtractZIP(archivePath, target); err == nil {
		t.Fatal("ExtractZIP accepted a regular file as target")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "not a directory" {
		t.Fatalf("target file was modified: %q", got)
	}
}

func TestExtractZIPRemovesFailedEntryButKeepsOtherTargetMaterial(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "corrupt.zip")
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "sentinel"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "candidate"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	writeTestZIP(t, archivePath, testEntry{name: "candidate", body: []byte("candidate data"), mode: 0755})
	corruptFirstEntryData(t, archivePath)

	err := ExtractZIP(archivePath, target)
	if err == nil || !strings.Contains(err.Error(), "candidate") {
		t.Fatalf("corrupt entry error = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, "candidate")); !errors.Is(statErr, fs.ErrNotExist) {
		t.Fatalf("failed overwritten candidate remains: %v", statErr)
	}
	got, readErr := os.ReadFile(filepath.Join(target, "sentinel"))
	if readErr != nil || string(got) != "keep" {
		t.Fatalf("unrelated target material changed: %q, %v", got, readErr)
	}
}

func TestExtractZIPRejectsPreexistingRelativeSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "symlink.zip")
	target := filepath.Join(root, "target")
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(target, "pivot")); err != nil {
		t.Skipf("cannot create directory symlink on this host: %v", err)
	}
	writeTestZIP(t, archivePath, testEntry{name: "pivot/escaped.txt", body: []byte("escaped"), mode: 0644})

	if err := ExtractZIP(archivePath, target); err == nil {
		t.Fatal("ExtractZIP followed a symlink outside target")
	}
	if _, err := os.Stat(filepath.Join(outside, "escaped.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive escaped through symlink: %v", err)
	}
}

func TestExtractZIPWritesSymlinkMetadataAsRegularFile(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "metadata-symlink.zip")
	target := filepath.Join(root, "target")
	writeTestZIP(t, archivePath, testEntry{name: "link", body: []byte("destination"), mode: fs.ModeSymlink | 0777})

	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(filepath.Join(target, "link"))
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("ZIP symlink metadata created mode %v", info.Mode())
	}
	got, err := os.ReadFile(filepath.Join(target, "link"))
	if err != nil || string(got) != "destination" {
		t.Fatalf("regular entry data = %q, %v", got, err)
	}
}

var (
	errEntryCopy        = errors.New("entry copy failed")
	errDestinationClose = errors.New("destination close failed")
	errSourceClose      = errors.New("source close failed")
)

type errorWriteCloser struct{}

func (errorWriteCloser) Write([]byte) (int, error) { return 0, errEntryCopy }
func (errorWriteCloser) Close() error              { return errDestinationClose }

type errorReadCloser struct{ io.Reader }

func (errorReadCloser) Close() error { return errSourceClose }

func TestCopyAndCloseJoinsCopyAndBothCloseErrors(t *testing.T) {
	err := copyAndClose(errorWriteCloser{}, errorReadCloser{Reader: strings.NewReader("data")})
	for _, want := range []error{errEntryCopy, errDestinationClose, errSourceClose} {
		if !errors.Is(err, want) {
			t.Fatalf("copyAndClose error %v does not contain %v", err, want)
		}
	}
}
