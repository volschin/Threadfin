//go:build !windows

package archive

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func fileModePerm(t *testing.T, name string) fs.FileMode {
	t.Helper()
	info, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func TestExtractZIPUsesRequestedModesSubjectToUmask(t *testing.T) {
	root := t.TempDir()
	referenceTarget := filepath.Join(root, "reference-target")
	referenceDirectory := filepath.Join(root, "reference-directory")
	referenceParent := filepath.Join(root, "reference-parent")
	referenceFile := filepath.Join(root, "reference-file")
	if err := os.Mkdir(referenceTarget, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(referenceDirectory, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(referenceParent, 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(referenceFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(root, "modes.zip")
	target := filepath.Join(root, "target")
	writeTestZIP(t, archivePath,
		testEntry{name: "bin/", mode: fs.ModeDir | 0750},
		testEntry{name: "bin/threadfin", body: []byte("binary"), mode: 0640},
		testEntry{name: "implicit/file", body: []byte("implicit"), mode: 0640},
		testEntry{name: "nested/explicit/", mode: fs.ModeDir | 0750},
	)
	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}

	checks := []struct{ got, reference string }{
		{got: target, reference: referenceTarget},
		{got: filepath.Join(target, "bin"), reference: referenceDirectory},
		{got: filepath.Join(target, "implicit"), reference: referenceParent},
		{got: filepath.Join(target, "nested"), reference: referenceParent},
		{got: filepath.Join(target, "nested", "explicit"), reference: referenceDirectory},
		{got: filepath.Join(target, "bin", "threadfin"), reference: referenceFile},
	}
	for _, check := range checks {
		if got, want := fileModePerm(t, check.got), fileModePerm(t, check.reference); got != want {
			t.Fatalf("%s mode = %#o, want reference mode %#o", check.got, got, want)
		}
	}
}

func TestExtractZIPPreservesExistingModeOnOverwrite(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "overwrite-mode.zip")
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(target, "threadfin")
	if err := os.WriteFile(existing, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	before := fileModePerm(t, existing)
	writeTestZIP(t, archivePath, testEntry{name: "threadfin", body: []byte("new"), mode: 0755})
	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	if after := fileModePerm(t, existing); after != before {
		t.Fatalf("overwrite changed mode from %#o to %#o", before, after)
	}
}

func TestExtractZIPTreatsBackslashAsUnixFilenameCharacter(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "backslash.zip")
	target := filepath.Join(root, "target")
	writeTestZIP(t, archivePath, testEntry{name: `dir\file.txt`, body: []byte("literal backslash"), mode: 0644})
	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(target, `dir\file.txt`))
	if err != nil || string(got) != "literal backslash" {
		t.Fatalf("backslash-named file = %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(target, "dir", "file.txt")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("backslash became a Unix separator: %v", err)
	}
}

func TestExtractZIPFollowsSymlinkThatStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "inside-link.zip")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Join(target, "real"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(target, "alias")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	writeTestZIP(t, archivePath, testEntry{name: "alias/file.txt", body: []byte("inside"), mode: 0644})
	if err := ExtractZIP(archivePath, target); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(target, "real", "file.txt"))
	if err != nil || string(got) != "inside" {
		t.Fatalf("in-root symlink result = %q, %v", got, err)
	}
}

func TestExtractZIPDoesNotEscapeDuringConcurrentSymlinkStress(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "race.zip")
	target := filepath.Join(root, "target")
	outside := filepath.Join(root, "outside")
	pivot := filepath.Join(target, "pivot")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestZIP(t, archivePath, testEntry{name: "pivot/escaped.txt", body: []byte("escaped"), mode: 0644})

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = os.RemoveAll(pivot)
			_ = os.Symlink("../outside", pivot)
			runtime.Gosched()
			_ = os.Remove(pivot)
			_ = os.Mkdir(pivot, 0755)
			runtime.Gosched()
		}
	}()
	defer func() {
		close(stop)
		<-done
	}()

	for range 100 {
		_ = ExtractZIP(archivePath, target)
		if _, err := os.Stat(filepath.Join(outside, "escaped.txt")); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("archive escaped during concurrent symlink replacement: %v", err)
		}
	}
}
