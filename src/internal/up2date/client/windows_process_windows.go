//go:build windows

package up2date

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

type osUpdateProcess struct {
	process   *os.Process
	mu        sync.Mutex
	finalized bool
}

func startOSUpdateProcess(executable string, args []string) (updateProcess, error) {
	if !filepath.IsAbs(executable) || filepath.Clean(executable) != executable {
		return nil, errors.New("update process executable must be absolute and clean")
	}
	argv := append([]string{executable}, args...)
	process, err := os.StartProcess(executable, argv, &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return nil, err
	}
	return &osUpdateProcess{process: process}, nil
}

func findOSUpdateProcess(pid int) (updateProcess, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	return &osUpdateProcess{process: process}, nil
}

func (process *osUpdateProcess) PID() int {
	return process.process.Pid
}

func (process *osUpdateProcess) Exited() (bool, error) {
	process.mu.Lock()
	defer process.mu.Unlock()
	if process.finalized {
		return false, errors.New("query finalized update process")
	}
	result, err := waitForWindowsProcess(process.process, 0)
	if err != nil {
		return false, err
	}
	return result == windows.WAIT_OBJECT_0, nil
}

func (process *osUpdateProcess) Kill() error {
	process.mu.Lock()
	defer process.mu.Unlock()
	if process.finalized {
		return errors.New("kill finalized update process")
	}
	return process.process.Kill()
}

func (process *osUpdateProcess) Wait(timeout time.Duration) error {
	process.mu.Lock()
	if process.finalized {
		process.mu.Unlock()
		return errors.New("wait finalized update process")
	}
	process.finalized = true
	process.mu.Unlock()

	result, waitErr := waitForWindowsProcess(process.process, timeout)
	if waitErr != nil {
		return errors.Join(waitErr, process.process.Release())
	}
	if result == uint32(windows.WAIT_TIMEOUT) {
		return errors.Join(errWindowsUpdateTimeout, process.process.Release())
	}
	_, err := process.process.Wait()
	return err
}

func (process *osUpdateProcess) Release() error {
	process.mu.Lock()
	if process.finalized {
		process.mu.Unlock()
		return errors.New("release finalized update process")
	}
	process.finalized = true
	process.mu.Unlock()
	return process.process.Release()
}

func waitForWindowsProcess(process *os.Process, timeout time.Duration) (uint32, error) {
	waitMilliseconds := uint32(0)
	if timeout > 0 {
		milliseconds := (timeout + time.Millisecond - 1) / time.Millisecond
		if milliseconds >= time.Duration(math.MaxUint32) {
			milliseconds = time.Duration(math.MaxUint32 - 1)
		}
		waitMilliseconds = uint32(milliseconds)
	}
	var result uint32
	var waitErr error
	err := process.WithHandle(func(handle uintptr) {
		result, waitErr = windows.WaitForSingleObject(windows.Handle(handle), waitMilliseconds)
	})
	if err != nil {
		return 0, err
	}
	if waitErr != nil {
		return 0, fmt.Errorf("wait for Windows process: %w", waitErr)
	}
	if result != windows.WAIT_OBJECT_0 && result != uint32(windows.WAIT_TIMEOUT) {
		return result, fmt.Errorf("unexpected Windows process wait result %d", result)
	}
	return result, nil
}
