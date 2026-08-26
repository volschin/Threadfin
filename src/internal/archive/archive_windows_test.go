//go:build windows

package archive

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractZIPClassifiesWindowsPaths(t *testing.T) {
	testCases := []struct {
		entry      string
		wantReason string
	}{
		{entry: `C:/escaped.txt`, wantReason: "is not local"},
		{entry: `NUL`, wantReason: "is not local"},
		{entry: `dir\file.txt`, wantReason: "is invalid on this platform"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.entry, func(t *testing.T) {
			root := t.TempDir()
			archivePath := filepath.Join(root, "invalid.zip")
			target := filepath.Join(root, "target")
			writeTestZIP(t, archivePath, testEntry{name: testCase.entry, body: []byte("invalid"), mode: 0644})
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
				t.Fatalf("rejected path created entries: %v", entries)
			}
		})
	}
}

func TestExtractZIPRejectsWindowsJunctionEscape(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "junction.zip")
	target := filepath.Join(root, "target")
	outside := filepath.Join(root, "outside")
	junction := filepath.Join(target, "pivot")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", junction, outside).CombinedOutput()
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("CI must create a Windows junction: %v: %s", err, output)
		}
		lowerOutput := strings.ToLower(string(output))
		if errors.Is(err, exec.ErrNotFound) ||
			strings.Contains(lowerOutput, "access is denied") ||
			strings.Contains(lowerOutput, "sufficient privilege") {
			t.Skipf("cannot create Windows junction on this host: %v: %s", err, output)
		}
		t.Fatalf("create Windows junction: %v: %s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove(junction) })
	writeTestZIP(t, archivePath, testEntry{name: "pivot/escaped.txt", body: []byte("escaped"), mode: 0644})

	if err := ExtractZIP(archivePath, target); err == nil {
		t.Fatal("ExtractZIP followed a junction outside target")
	}
	if _, err := os.Stat(filepath.Join(outside, "escaped.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive escaped through junction: %v", err)
	}
}
