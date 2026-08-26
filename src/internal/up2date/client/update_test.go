package up2date

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRestorOldBinaryReturnsRenameError(t *testing.T) {
	directory := t.TempDir()
	newBinary := filepath.Join(directory, "threadfin")
	if err := os.WriteFile(newBinary, []byte("failed update"), 0755); err != nil {
		t.Fatal(err)
	}

	err := restorOldBinary(filepath.Join(directory, "missing-old-threadfin"), newBinary)
	if err == nil {
		t.Fatal("missing backup restoration succeeded")
	}
	if _, ok := errors.AsType[*os.LinkError](err); !ok {
		t.Fatalf("restoration error = %v, want filesystem error", err)
	}
}

func TestReplacePreparedUpdateCopyFailureRestoresCurrentBinary(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin")
	oldBinary := filepath.Join(directory, "old-threadfin")
	if err := os.WriteFile(binary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}

	err := replacePreparedUpdate(filepath.Join(directory, "missing-candidate"), binary, oldBinary)
	if err == nil {
		t.Fatal("missing candidate was installed")
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after copy failure", got)
	}
}

func TestReplacePreparedUpdateChmodFailureReportsRollbackFailure(t *testing.T) {
	directory := t.TempDir()
	candidate := filepath.Join(directory, "candidate")
	binary := filepath.Join(directory, "threadfin")
	oldBinary := filepath.Join(directory, "old-threadfin")
	if err := os.WriteFile(candidate, []byte("new executable"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}

	originalChmod := chmodUpdateBinary
	t.Cleanup(func() { chmodUpdateBinary = originalChmod })
	chmodErr := errors.New("chmod failed")
	chmodUpdateBinary = func(string, os.FileMode) error {
		if err := os.Remove(oldBinary); err != nil {
			return err
		}
		return chmodErr
	}

	err := replacePreparedUpdate(candidate, binary, oldBinary)
	if !errors.Is(err, chmodErr) {
		t.Fatalf("replacement error = %v, want chmod error", err)
	}
	if _, ok := errors.AsType[*os.LinkError](err); !ok {
		t.Fatalf("replacement error = %v, want rollback filesystem error", err)
	}
}

func TestRestartUnixCleansCandidateBeforeSuccessfulExec(t *testing.T) {
	candidate := filepath.Join(t.TempDir(), "candidate")
	if err := os.WriteFile(candidate, []byte("verified update"), 0600); err != nil {
		t.Fatal(err)
	}
	execCalled := false
	execProcess := func(string, []string, []string) error {
		execCalled = true
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			return errors.New("candidate remained when exec was called")
		}
		return nil
	}

	err := restartUnix("threadfin", "old-threadfin", func() error { return os.Remove(candidate) }, execProcess)
	if err != nil {
		t.Fatal(err)
	}
	if !execCalled {
		t.Fatal("exec was not called")
	}
}

func TestRestartUnixReportsExecAndRollbackFailures(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	execErr := errors.New("exec failed")
	execProcess := func(string, []string, []string) error { return execErr }

	err := restartUnix(binary, filepath.Join(directory, "missing-old-threadfin"), func() error { return nil }, execProcess)
	if !errors.Is(err, execErr) {
		t.Fatalf("restart error = %v, want exec error", err)
	}
	if _, ok := errors.AsType[*os.LinkError](err); !ok {
		t.Fatalf("restart error = %v, want rollback filesystem error", err)
	}
}
