package up2date

import (
	"archive/zip"
	"errors"
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
	var linkErr *os.LinkError
	if !errors.As(err, &linkErr) {
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
	var linkErr *os.LinkError
	if !errors.As(err, &linkErr) {
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
	var linkErr *os.LinkError
	if !errors.As(err, &linkErr) {
		t.Fatalf("restart error = %v, want rollback filesystem error", err)
	}
}

func TestRestartWindowsReturnsStartErrorAfterRestoringBinary(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	startErr := errors.New("start failed")
	startProcess := func(...string) (*os.Process, error) { return nil, startErr }
	findProcess := func(int) (*os.Process, error) { return &os.Process{}, nil }
	killProcess := func(*os.Process) error { return errors.New("kill should not be called") }
	waitProcess := func(*os.Process) error { return errors.New("wait should not be called") }

	err := restartWindows(binary, oldBinary, func() error { return nil }, os.RemoveAll, findProcess, startProcess, killProcess, waitProcess)
	if !errors.Is(err, startErr) {
		t.Fatalf("restart error = %v, want start error", err)
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after start failure", got)
	}
}

func TestRestartWindowsCleansCandidateBeforeSuccessfulCutover(t *testing.T) {
	directory := t.TempDir()
	candidate := filepath.Join(directory, "candidate")
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	for path, body := range map[string]string{
		candidate: "verified candidate",
		binary:    "new executable",
		oldBinary: "current executable",
	} {
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}

	currentProcess := &os.Process{}
	updatedProcess := &os.Process{}
	started := false
	killed := false
	waited := false
	cleanup := func() error { return os.Remove(candidate) }
	findProcess := func(pid int) (*os.Process, error) {
		if pid != os.Getpid() {
			t.Fatalf("find process PID = %d, want %d", pid, os.Getpid())
		}
		return currentProcess, nil
	}
	startProcess := func(args ...string) (*os.Process, error) {
		started = true
		if len(args) != 1 || args[0] != binary {
			t.Fatalf("start args = %v, want [%s]", args, binary)
		}
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			t.Fatalf("candidate remains when updated process starts: %v", err)
		}
		return updatedProcess, nil
	}
	killProcess := func(process *os.Process) error {
		killed = true
		if process != currentProcess {
			t.Fatalf("killed process = %p, want current process %p", process, currentProcess)
		}
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			t.Fatalf("candidate remains when current process is killed: %v", err)
		}
		return nil
	}
	waitProcess := func(process *os.Process) error {
		waited = true
		if process != updatedProcess {
			t.Fatalf("waited process = %p, want updated process %p", process, updatedProcess)
		}
		return nil
	}

	if err := restartWindows(binary, oldBinary, cleanup, os.RemoveAll, findProcess, startProcess, killProcess, waitProcess); err != nil {
		t.Fatal(err)
	}
	if !started || !killed || !waited {
		t.Fatalf("cutover calls: started=%t killed=%t waited=%t", started, killed, waited)
	}
	if _, err := os.Stat(oldBinary); !os.IsNotExist(err) {
		t.Fatalf("previous executable remains after successful cutover: %v", err)
	}
}

func TestRestartWindowsCleanupFailureRestoresBinary(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	cleanupErr := errors.New("cleanup failed")
	startCalled := false
	startProcess := func(...string) (*os.Process, error) {
		startCalled = true
		return nil, nil
	}

	err := restartWindows(
		binary,
		oldBinary,
		func() error { return cleanupErr },
		os.RemoveAll,
		func(int) (*os.Process, error) { return &os.Process{}, nil },
		startProcess,
		func(*os.Process) error { return nil },
		func(*os.Process) error { return nil },
	)
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("restart error = %v, want cleanup error", err)
	}
	if startCalled {
		t.Fatal("updated process started after cleanup failure")
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after cleanup failure", got)
	}
}

func TestRestartWindowsReturnsKillError(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	killErr := errors.New("kill failed")
	currentProcess := &os.Process{}
	updatedProcess := &os.Process{}
	replacementStopped := false
	replacementReaped := false
	killProcess := func(process *os.Process) error {
		switch process {
		case currentProcess:
			return killErr
		case updatedProcess:
			replacementStopped = true
			return nil
		default:
			t.Fatalf("unexpected process killed: %p", process)
			return nil
		}
	}
	waitProcess := func(process *os.Process) error {
		if process != updatedProcess {
			t.Fatalf("waited process = %p, want updated process %p", process, updatedProcess)
		}
		replacementReaped = true
		return nil
	}

	err := restartWindows(
		binary,
		oldBinary,
		func() error { return nil },
		os.RemoveAll,
		func(int) (*os.Process, error) { return currentProcess, nil },
		func(...string) (*os.Process, error) { return updatedProcess, nil },
		killProcess,
		waitProcess,
	)
	if !errors.Is(err, killErr) {
		t.Fatalf("restart error = %v, want kill error", err)
	}
	if !replacementStopped || !replacementReaped {
		t.Fatalf("replacement cleanup: stopped=%t reaped=%t", replacementStopped, replacementReaped)
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after kill failure", got)
	}
	if _, err := os.Stat(oldBinary); !os.IsNotExist(err) {
		t.Fatalf("rollback binary remains after restoration: %v", err)
	}
}

func TestRestartWindowsReturnsWaitError(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	waitErr := errors.New("wait failed")
	currentProcess := &os.Process{}
	updatedProcess := &os.Process{}
	replacementStopped := false
	replacementReaped := false
	waitCalls := 0
	killProcess := func(process *os.Process) error {
		switch process {
		case currentProcess:
			return nil
		case updatedProcess:
			replacementStopped = true
			return nil
		default:
			t.Fatalf("unexpected process killed: %p", process)
			return nil
		}
	}
	waitProcess := func(process *os.Process) error {
		if process != updatedProcess {
			t.Fatalf("waited process = %p, want updated process %p", process, updatedProcess)
		}
		waitCalls++
		if waitCalls == 1 {
			return waitErr
		}
		replacementReaped = true
		return nil
	}

	err := restartWindows(
		binary,
		oldBinary,
		func() error { return nil },
		os.RemoveAll,
		func(int) (*os.Process, error) { return currentProcess, nil },
		func(...string) (*os.Process, error) { return updatedProcess, nil },
		killProcess,
		waitProcess,
	)
	if !errors.Is(err, waitErr) {
		t.Fatalf("restart error = %v, want wait error", err)
	}
	if !replacementStopped || !replacementReaped {
		t.Fatalf("replacement cleanup: stopped=%t reaped=%t", replacementStopped, replacementReaped)
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after wait failure", got)
	}
	if _, err := os.Stat(oldBinary); !os.IsNotExist(err) {
		t.Fatalf("rollback binary remains after restoration: %v", err)
	}
}

func TestRestartWindowsReplacementStopFailureRetainsRollbackState(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	killErr := errors.New("kill current failed")
	stopErr := errors.New("stop replacement failed")
	currentProcess := &os.Process{}
	updatedProcess := &os.Process{}
	killProcess := func(process *os.Process) error {
		switch process {
		case currentProcess:
			return killErr
		case updatedProcess:
			return stopErr
		default:
			t.Fatalf("unexpected process killed: %p", process)
			return nil
		}
	}

	err := restartWindows(
		binary,
		oldBinary,
		func() error { return nil },
		os.RemoveAll,
		func(int) (*os.Process, error) { return currentProcess, nil },
		func(...string) (*os.Process, error) { return updatedProcess, nil },
		killProcess,
		func(*os.Process) error {
			t.Fatal("replacement wait called after stop failure")
			return nil
		},
	)
	if !errors.Is(err, killErr) || !errors.Is(err, stopErr) {
		t.Fatalf("restart error = %v, want current-kill and replacement-stop errors", err)
	}
	gotBinary, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(gotBinary) != "new executable" {
		t.Fatalf("new executable = %q after replacement stop failure", gotBinary)
	}
	gotRollback, readErr := os.ReadFile(oldBinary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(gotRollback) != "current executable" {
		t.Fatalf("rollback executable = %q after replacement stop failure", gotRollback)
	}
}

func TestRestartWindowsOldBinaryCleanupFailureRestoresBinary(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "threadfin.exe")
	oldBinary := filepath.Join(directory, "old-threadfin.exe")
	if err := os.WriteFile(binary, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, []byte("current executable"), 0755); err != nil {
		t.Fatal(err)
	}
	cleanupErr := errors.New("old-binary cleanup failed")
	currentProcess := &os.Process{}
	updatedProcess := &os.Process{}
	currentStopped := false
	replacementReaped := false
	removeOldBinary := func(path string) error {
		if path != oldBinary {
			t.Fatalf("old-binary cleanup path = %q, want %q", path, oldBinary)
		}
		return cleanupErr
	}
	killProcess := func(process *os.Process) error {
		if process != currentProcess {
			t.Fatalf("unexpected process killed: %p", process)
		}
		currentStopped = true
		return nil
	}
	waitProcess := func(process *os.Process) error {
		if process != updatedProcess {
			t.Fatalf("waited process = %p, want updated process %p", process, updatedProcess)
		}
		replacementReaped = true
		return nil
	}

	err := restartWindows(
		binary,
		oldBinary,
		func() error { return nil },
		removeOldBinary,
		func(int) (*os.Process, error) { return currentProcess, nil },
		func(...string) (*os.Process, error) { return updatedProcess, nil },
		killProcess,
		waitProcess,
	)
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("restart error = %v, want old-binary cleanup error", err)
	}
	if !currentStopped || !replacementReaped {
		t.Fatalf("cutover state: currentStopped=%t replacementReaped=%t", currentStopped, replacementReaped)
	}
	got, readErr := os.ReadFile(binary)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "current executable" {
		t.Fatalf("current executable = %q after old-binary cleanup failure", got)
	}
	if _, err := os.Stat(oldBinary); !os.IsNotExist(err) {
		t.Fatalf("rollback binary remains after restoration: %v", err)
	}
}
