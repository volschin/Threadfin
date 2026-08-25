package src

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/avfs/avfs/vfs/memfs"
)

func TestTerminateProcessKillsAndReapsChild(t *testing.T) {
	if os.Getenv("THREADFIN_PROCESS_HELPER") == "1" {
		time.Sleep(time.Minute)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestTerminateProcessKillsAndReapsChild$")
	cmd.Env = append(os.Environ(), "THREADFIN_PROCESS_HELPER=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	if err := terminateProcess(cmd); err != nil {
		t.Fatalf("terminateProcess() error = %v", err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("terminateProcess() did not reap the child process")
	}
}

func TestCreateBufferFileReturnsCreateError(t *testing.T) {
	previousVFS := bufferVFS
	bufferVFS = memfs.New()
	t.Cleanup(func() { bufferVFS = previousVFS })

	err := createBufferFile(filepath.Join("missing", "segment.ts"))
	if err == nil {
		t.Fatal("createBufferFile() error = nil, want missing-directory error")
	}
}
