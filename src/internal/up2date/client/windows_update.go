package up2date

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// MinimumWindowsUpdateReadinessTimeout is the bounded allowance for local
	// initialization after configured sequential provider waits are added.
	MinimumWindowsUpdateReadinessTimeout = 2 * time.Minute

	windowsUpdateStateVersion = 1
	windowsUpdateNonceBytes   = 32
	windowsUpdateStateLimit   = 128 << 10

	windowsUpdateHelperMode = "--threadfin-private-update-helper-v1"
	windowsUpdateChildMode  = "--threadfin-private-update-child-v1"

	windowsUpdateAckMarker      = "ACK"
	windowsUpdateReadyMarker    = "READY"
	windowsUpdateCompleteMarker = "COMPLETE"

	windowsUpdateAckTimeout       = 10 * time.Second
	windowsOldProcessExitTimeout  = 30 * time.Second
	windowsChildReadinessTimeout  = MinimumWindowsUpdateReadinessTimeout
	windowsProcessStopTimeout     = 10 * time.Second
	windowsHelperExitTimeout      = 30 * time.Second
	windowsUpdatePollInterval     = 50 * time.Millisecond
	windowsCompletionRetryTimeout = time.Second
)

var (
	ErrWindowsUpdateHandoff = errors.New("Windows update helper acknowledged handoff")
	errWindowsUpdateTimeout = errors.New("Windows update condition timed out")
	errProcessUncertain     = errors.New("Windows update process state is uncertain")
)

type windowsUpdateAttempt string

const (
	windowsReplacementAttempt windowsUpdateAttempt = "replacement"
	windowsRecoveryAttempt    windowsUpdateAttempt = "recovery"
)

type windowsUpdateState struct {
	Version          int           `json:"version"`
	Nonce            string        `json:"nonce"`
	Canonical        string        `json:"canonical"`
	Backup           string        `json:"backup"`
	Args             []string      `json:"args"`
	OldPID           int           `json:"old_pid"`
	ReadinessTimeout time.Duration `json:"readiness_timeout,omitempty"`
}

type updateProcess interface {
	PID() int
	Exited() (bool, error)
	Kill() error
	Wait(time.Duration) error
	Release() error
}

type windowsUpdateProtocol struct {
	startProcess     func(string, []string) (updateProcess, error)
	findProcess      func(int) (updateProcess, error)
	currentPID       func() int
	executable       func() (string, error)
	pollCondition    func(time.Duration, func() (bool, error)) (bool, error)
	readinessTimeout time.Duration
	writeMarker      func(string, string, string) error
}

type windowsUpdateMarkerOps struct {
	closeFile     func(*os.File) error
	beforePublish func()
	publish       func(string, string) error
}

type windowsUpdateChild struct {
	state        windowsUpdateState
	statePath    string
	attempt      windowsUpdateAttempt
	helperPID    int
	originalArgs []string
}

type privateUpdateStartup struct {
	private             bool
	exit                bool
	exitCode            int
	originalArgs        []string
	skipAutomaticUpdate bool
}

var activeWindowsUpdateChild struct {
	sync.Mutex
	child    *windowsUpdateChild
	signaled bool
}

func defaultWindowsUpdateProtocol() windowsUpdateProtocol {
	return windowsUpdateProtocol{
		startProcess:     startOSUpdateProcess,
		findProcess:      findOSUpdateProcess,
		currentPID:       os.Getpid,
		executable:       os.Executable,
		pollCondition:    pollWindowsUpdateCondition,
		readinessTimeout: normalizedWindowsUpdateReadinessTimeout(Updater.WindowsUpdateReadinessTimeout),
		writeMarker:      writeWindowsUpdateMarker,
	}
}

func newWindowsUpdateState(canonical, backup string, args []string, oldPID int, readinessTimeout time.Duration) (windowsUpdateState, string, error) {
	nonceBytes := make([]byte, windowsUpdateNonceBytes)
	if _, err := rand.Read(nonceBytes); err != nil {
		return windowsUpdateState{}, "", fmt.Errorf("generate update nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	state := windowsUpdateState{
		Version:          windowsUpdateStateVersion,
		Nonce:            nonce,
		Canonical:        filepath.Clean(canonical),
		Backup:           filepath.Clean(backup),
		Args:             append([]string{}, args...),
		OldPID:           oldPID,
		ReadinessTimeout: normalizedWindowsUpdateReadinessTimeout(readinessTimeout),
	}
	statePath := windowsUpdateStatePath(state.Canonical, nonce)
	if err := validateWindowsUpdateState(state, statePath, nonce, state.Backup); err != nil {
		return windowsUpdateState{}, "", err
	}
	return state, statePath, nil
}

func windowsUpdateStatePath(canonical, nonce string) string {
	return canonical + ".update-" + nonce + ".json"
}

func windowsUpdateAckPath(statePath string) string {
	return statePath + ".ack"
}

func windowsUpdateReadyPath(statePath string, attempt windowsUpdateAttempt) string {
	return statePath + "." + string(attempt) + ".ready"
}

func windowsUpdateCompletionPath(statePath string) string {
	return statePath + ".complete"
}

func validateWindowsUpdateState(state windowsUpdateState, statePath, nonce, executable string) error {
	if state.Version != windowsUpdateStateVersion {
		return fmt.Errorf("unsupported Windows update state version %d", state.Version)
	}
	if err := validateWindowsUpdateNonce(nonce); err != nil {
		return err
	}
	if state.Nonce != nonce {
		return errors.New("Windows update nonce does not match state")
	}
	for name, path := range map[string]string{
		"canonical executable": state.Canonical,
		"backup executable":    state.Backup,
		"state file":           statePath,
		"running executable":   executable,
	} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return fmt.Errorf("%s path is not absolute and clean", name)
		}
	}
	expectedBackup := filepath.Join(filepath.Dir(state.Canonical), "_old_"+filepath.Base(state.Canonical))
	if !sameUpdatePath(state.Backup, expectedBackup) {
		return errors.New("Windows update backup is not the canonical executable's known-good generation")
	}
	if !sameUpdatePath(statePath, windowsUpdateStatePath(state.Canonical, nonce)) {
		return errors.New("Windows update state file is outside the canonical relationship")
	}
	if !sameUpdatePath(executable, state.Canonical) && !sameUpdatePath(executable, state.Backup) {
		return errors.New("running executable is not a permitted update generation")
	}
	if state.OldPID <= 0 {
		return errors.New("Windows update state has an invalid old PID")
	}
	if state.Args == nil {
		return errors.New("Windows update state has no original argument vector")
	}
	if state.ReadinessTimeout != 0 && state.ReadinessTimeout < windowsChildReadinessTimeout {
		return errors.New("Windows update readiness budget is below the protocol minimum")
	}
	return nil
}

func normalizedWindowsUpdateReadinessTimeout(timeout time.Duration) time.Duration {
	// A zero value keeps state written by the first helper protocol readable.
	if timeout < windowsChildReadinessTimeout {
		return windowsChildReadinessTimeout
	}
	return timeout
}

func validateWindowsUpdateNonce(nonce string) error {
	if len(nonce) != hex.EncodedLen(windowsUpdateNonceBytes) {
		return errors.New("Windows update nonce has invalid length")
	}
	decoded, err := hex.DecodeString(nonce)
	if err != nil || len(decoded) != windowsUpdateNonceBytes || hex.EncodeToString(decoded) != nonce {
		return errors.New("Windows update nonce is not canonical random hex")
	}
	return nil
}

func sameUpdatePath(left, right string) bool {
	// This comparator is only for the Windows update protocol. Keeping its
	// semantics independent of the test host makes casing regressions testable.
	return strings.EqualFold(left, right)
}

func writeWindowsUpdateState(statePath string, state windowsUpdateState) (err error) {
	if err := validateWindowsUpdateState(state, statePath, state.Nonce, state.Backup); err != nil {
		return err
	}
	file, err := os.OpenFile(statePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create Windows update state: %w", err)
	}
	keep := false
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close Windows update state: %w", closeErr)
		}
		if !keep || err != nil {
			_ = os.Remove(statePath)
		}
	}()
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(state); err != nil {
		return fmt.Errorf("encode Windows update state: %w", err)
	}
	if err = file.Sync(); err != nil {
		return fmt.Errorf("sync Windows update state: %w", err)
	}
	keep = true
	return nil
}

func readWindowsUpdateState(statePath, nonce, executable string) (windowsUpdateState, error) {
	file, err := os.Open(statePath)
	if err != nil {
		return windowsUpdateState{}, fmt.Errorf("open Windows update state: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return windowsUpdateState{}, fmt.Errorf("stat Windows update state: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() > windowsUpdateStateLimit {
		return windowsUpdateState{}, errors.New("Windows update state is not a bounded regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return windowsUpdateState{}, errors.New("Windows update state is not owner-only")
	}
	decoder := json.NewDecoder(io.LimitReader(file, windowsUpdateStateLimit+1))
	decoder.DisallowUnknownFields()
	var state windowsUpdateState
	if err := decoder.Decode(&state); err != nil {
		return windowsUpdateState{}, fmt.Errorf("decode Windows update state: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return windowsUpdateState{}, errors.New("Windows update state has trailing data")
	}
	if err := validateWindowsUpdateState(state, statePath, nonce, executable); err != nil {
		return windowsUpdateState{}, err
	}
	return state, nil
}

func defaultWindowsUpdateMarkerOps() windowsUpdateMarkerOps {
	return windowsUpdateMarkerOps{
		closeFile: func(file *os.File) error { return file.Close() },
		publish:   os.Link,
	}
}

func writeWindowsUpdateMarker(path, nonce, kind string) error {
	return writeWindowsUpdateMarkerWithOps(path, nonce, kind, defaultWindowsUpdateMarkerOps())
}

func writeWindowsUpdateMarkerWithOps(path, nonce, kind string, ops windowsUpdateMarkerOps) error {
	if err := validateWindowsUpdateNonce(nonce); err != nil {
		return err
	}
	if kind != windowsUpdateAckMarker && kind != windowsUpdateReadyMarker && kind != windowsUpdateCompleteMarker {
		return errors.New("invalid Windows update marker kind")
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary Windows update marker: %w", err)
	}
	temporaryPath := file.Name()
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
		_ = os.Remove(temporaryPath)
	}()
	if _, err := io.WriteString(file, kind+":"+nonce+"\n"); err != nil {
		return fmt.Errorf("write Windows update marker: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync Windows update marker: %w", err)
	}
	if err := ops.closeFile(file); err != nil {
		return fmt.Errorf("close Windows update marker: %w", err)
	}
	closed = true
	if ops.beforePublish != nil {
		ops.beforePublish()
	}
	if err := ops.publish(temporaryPath, path); err != nil {
		return fmt.Errorf("publish Windows update marker: %w", err)
	}
	return nil
}

func windowsUpdateMarkerExists(path, nonce, kind string) (bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("open Windows update marker: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("stat Windows update marker: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() > 256 {
		return false, errors.New("Windows update marker is not a bounded regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return false, errors.New("Windows update marker is not owner-only")
	}
	data, err := io.ReadAll(io.LimitReader(file, 257))
	if err != nil {
		return false, fmt.Errorf("read Windows update marker: %w", err)
	}
	if string(data) != kind+":"+nonce+"\n" {
		return false, errors.New("Windows update marker does not match its nonce and phase")
	}
	return true, nil
}

func pollWindowsUpdateCondition(timeout time.Duration, condition func() (bool, error)) (bool, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(windowsUpdatePollInterval)
	defer ticker.Stop()
	for {
		done, err := condition()
		if done || err != nil {
			return done, err
		}
		select {
		case <-deadline.C:
			return false, nil
		case <-ticker.C:
		}
	}
}

func beginWindowsHandoff(canonical, backup string, originalArgs []string, cleanup func() error, protocol windowsUpdateProtocol) error {
	if err := cleanup(); err != nil {
		return restoreAfterUpdateFailure(fmt.Errorf("clean temporary update material: %w", err), backup, canonical)
	}
	state, statePath, err := newWindowsUpdateState(canonical, backup, originalArgs, protocol.currentPID(), protocol.readinessTimeout)
	if err != nil {
		return restoreAfterUpdateFailure(err, backup, canonical)
	}
	if err := writeWindowsUpdateState(statePath, state); err != nil {
		return restoreAfterUpdateFailure(err, backup, canonical)
	}
	helper, err := protocol.startProcess(state.Backup, []string{windowsUpdateHelperMode, statePath, state.Nonce})
	if err != nil {
		return rollbackBeforeWindowsHandoff(err, state, statePath)
	}
	helperExited := false
	acknowledged, pollErr := protocol.pollCondition(windowsUpdateAckTimeout, func() (bool, error) {
		acknowledged, err := windowsUpdateMarkerExists(windowsUpdateAckPath(statePath), state.Nonce, windowsUpdateAckMarker)
		if err != nil {
			return false, err
		}
		helperExited, err = helper.Exited()
		if helperExited || err != nil {
			return helperExited, err
		}
		return acknowledged, nil
	})
	if pollErr == nil && acknowledged && !helperExited {
		releaseErr := helper.Release()
		return errors.Join(ErrWindowsUpdateHandoff, releaseErr)
	}
	var stopErr error
	confirmedStopped := helperExited
	if helperExited {
		stopErr = helper.Wait(windowsProcessStopTimeout)
	} else {
		confirmedStopped, stopErr = stopAndWaitWindowsUpdateProcess(helper)
	}
	cause := pollErr
	if cause == nil {
		cause = errWindowsUpdateTimeout
	}
	cause = errors.Join(cause, stopErr)
	if !confirmedStopped {
		return errors.Join(errProcessUncertain, cause)
	}
	return rollbackBeforeWindowsHandoff(cause, state, statePath)
}

func rollbackBeforeWindowsHandoff(cause error, state windowsUpdateState, statePath string) error {
	restoreErr := restorOldBinary(state.Backup, state.Canonical)
	cleanupErr := removeWindowsUpdateMetadata(statePath)
	return errors.Join(cause, wrapError("restore previous binary", restoreErr), wrapError("remove Windows update state", cleanupErr))
}

func runWindowsUpdateHelper(statePath, nonce string, protocol windowsUpdateProtocol) error {
	executable, err := protocol.executable()
	if err != nil {
		return fmt.Errorf("resolve helper executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return fmt.Errorf("make helper executable absolute: %w", err)
	}
	state, err := readWindowsUpdateState(statePath, nonce, filepath.Clean(executable))
	if err != nil {
		return err
	}
	if !sameUpdatePath(filepath.Clean(executable), state.Backup) {
		return errors.New("Windows update helper is not running from the known-good backup")
	}
	oldProcess, err := protocol.findProcess(state.OldPID)
	if err != nil {
		return fmt.Errorf("open old Threadfin process: %w", err)
	}
	if err := writeWindowsUpdateMarker(windowsUpdateAckPath(statePath), state.Nonce, windowsUpdateAckMarker); err != nil {
		return errors.Join(err, oldProcess.Release())
	}
	if err := oldProcess.Wait(windowsOldProcessExitTimeout); err != nil {
		return fmt.Errorf("wait for old Threadfin process: %w", err)
	}

	replacementErr := runWindowsUpdateAttempt(state, statePath, windowsReplacementAttempt, protocol)
	if replacementErr == nil {
		return nil
	}
	if errors.Is(replacementErr, errProcessUncertain) {
		return replacementErr
	}
	if err := restoreKnownGoodCopy(state.Backup, state.Canonical); err != nil {
		return errors.Join(replacementErr, fmt.Errorf("restore known-good executable: %w", err))
	}
	recoveryErr := runWindowsUpdateAttempt(state, statePath, windowsRecoveryAttempt, protocol)
	if recoveryErr != nil {
		return errors.Join(replacementErr, fmt.Errorf("start restored Threadfin: %w", recoveryErr))
	}
	log.Printf("Windows update replacement failed and known-good Threadfin was restored: %v", replacementErr)
	return nil
}

func runWindowsUpdateAttempt(state windowsUpdateState, statePath string, attempt windowsUpdateAttempt, protocol windowsUpdateProtocol) error {
	if err := validateWindowsUpdateAttempt(attempt); err != nil {
		return err
	}
	readyPath := windowsUpdateReadyPath(statePath, attempt)
	if err := removeFileIfExists(readyPath); err != nil {
		return fmt.Errorf("clear previous %s readiness marker: %w", attempt, err)
	}
	args := []string{
		windowsUpdateChildMode,
		statePath,
		state.Nonce,
		string(attempt),
		strconv.Itoa(protocol.currentPID()),
	}
	child, err := protocol.startProcess(state.Canonical, args)
	if err != nil {
		return fmt.Errorf("start %s Threadfin: %w", attempt, err)
	}
	return awaitWindowsUpdateReadiness(child, state, statePath, attempt, protocol)
}

func awaitWindowsUpdateReadiness(child updateProcess, state windowsUpdateState, statePath string, attempt windowsUpdateAttempt, protocol windowsUpdateProtocol) error {
	childExited := false
	ready := false
	done, pollErr := protocol.pollCondition(normalizedWindowsUpdateReadinessTimeout(state.ReadinessTimeout), func() (bool, error) {
		var err error
		ready, err = windowsUpdateMarkerExists(windowsUpdateReadyPath(statePath, attempt), state.Nonce, windowsUpdateReadyMarker)
		if err != nil {
			return false, err
		}
		childExited, err = child.Exited()
		if childExited || err != nil {
			return childExited, err
		}
		return ready, nil
	})
	if pollErr == nil && done && ready && !childExited {
		if err := child.Release(); err != nil {
			return errors.Join(errProcessUncertain, fmt.Errorf("release ready %s process: %w", attempt, err))
		}
		return nil
	}
	if childExited {
		waitErr := child.Wait(windowsProcessStopTimeout)
		return errors.Join(fmt.Errorf("%s Threadfin exited before listener readiness", attempt), pollErr, waitErr)
	}
	cause := pollErr
	if cause == nil {
		cause = fmt.Errorf("%s Threadfin readiness: %w", attempt, errWindowsUpdateTimeout)
	}
	stopped, stopErr := stopAndWaitWindowsUpdateProcess(child)
	if !stopped {
		return errors.Join(errProcessUncertain, cause, stopErr)
	}
	return errors.Join(cause, stopErr)
}

func stopAndWaitWindowsUpdateProcess(process updateProcess) (bool, error) {
	killErr := process.Kill()
	waitErr := process.Wait(windowsProcessStopTimeout)
	return waitErr == nil, errors.Join(wrapError("kill update process", killErr), wrapError("wait for killed update process", waitErr))
}

func loadWindowsUpdateChild(statePath, nonce string, attempt windowsUpdateAttempt, helperPID int, executable string) (*windowsUpdateChild, error) {
	if err := validateWindowsUpdateAttempt(attempt); err != nil {
		return nil, err
	}
	if helperPID <= 0 {
		return nil, errors.New("Windows update child has invalid helper PID")
	}
	executable, err := filepath.Abs(executable)
	if err != nil {
		return nil, fmt.Errorf("make child executable absolute: %w", err)
	}
	state, err := readWindowsUpdateState(statePath, nonce, filepath.Clean(executable))
	if err != nil {
		return nil, err
	}
	if !sameUpdatePath(filepath.Clean(executable), state.Canonical) {
		return nil, errors.New("Windows update child is not running from the canonical executable")
	}
	if helperPID == state.OldPID {
		return nil, errors.New("Windows update child helper PID matches the old process")
	}
	return &windowsUpdateChild{
		state:        state,
		statePath:    statePath,
		attempt:      attempt,
		helperPID:    helperPID,
		originalArgs: append([]string{}, state.Args...),
	}, nil
}

func validateWindowsUpdateAttempt(attempt windowsUpdateAttempt) error {
	if attempt != windowsReplacementAttempt && attempt != windowsRecoveryAttempt {
		return fmt.Errorf("invalid Windows update attempt %q", attempt)
	}
	return nil
}

func finishWindowsUpdateChild(state windowsUpdateState, statePath string, helper updateProcess, protocol windowsUpdateProtocol) error {
	if err := helper.Wait(windowsHelperExitTimeout); err != nil {
		return fmt.Errorf("wait for Windows update helper: %w", err)
	}
	if err := publishWindowsUpdateCompletion(statePath, state.Nonce, protocol); err != nil {
		return fmt.Errorf("publish Windows update completion: %w", err)
	}
	return cleanupCompletedWindowsUpdate(state, statePath)
}

func publishWindowsUpdateCompletion(statePath, nonce string, protocol windowsUpdateProtocol) error {
	completionPath := windowsUpdateCompletionPath(statePath)
	var lastWriteErr error
	published, pollErr := protocol.pollCondition(windowsCompletionRetryTimeout, func() (bool, error) {
		exists, err := windowsUpdateMarkerExists(completionPath, nonce, windowsUpdateCompleteMarker)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
		lastWriteErr = protocol.writeMarker(completionPath, nonce, windowsUpdateCompleteMarker)
		return lastWriteErr == nil, nil
	})
	if pollErr != nil {
		return errors.Join(pollErr, lastWriteErr)
	}
	if !published {
		return errors.Join(errWindowsUpdateTimeout, lastWriteErr)
	}
	return nil
}

func cleanupCompletedWindowsUpdate(state windowsUpdateState, statePath string) error {
	if err := removeFileIfExists(state.Backup); err != nil {
		return fmt.Errorf("remove known-good update backup: %w", err)
	}
	for _, marker := range []string{
		windowsUpdateAckPath(statePath),
		windowsUpdateReadyPath(statePath, windowsReplacementAttempt),
		windowsUpdateReadyPath(statePath, windowsRecoveryAttempt),
	} {
		if err := removeFileIfExists(marker); err != nil {
			return fmt.Errorf("remove Windows update marker: %w", err)
		}
	}
	if err := removeFileIfExists(statePath); err != nil {
		return fmt.Errorf("remove Windows update state: %w", err)
	}
	if err := removeFileIfExists(windowsUpdateCompletionPath(statePath)); err != nil {
		return fmt.Errorf("remove Windows update completion marker: %w", err)
	}
	return nil
}

func removeWindowsUpdateMetadata(statePath string) error {
	var result error
	for _, path := range []string{
		windowsUpdateAckPath(statePath),
		windowsUpdateReadyPath(statePath, windowsReplacementAttempt),
		windowsUpdateReadyPath(statePath, windowsRecoveryAttempt),
		windowsUpdateCompletionPath(statePath),
		statePath,
	} {
		result = errors.Join(result, removeFileIfExists(path))
	}
	return result
}

func removeFileIfExists(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func restoreKnownGoodCopy(backup, canonical string) (err error) {
	input, err := os.Open(backup)
	if err != nil {
		return err
	}
	defer input.Close()
	temporary, err := os.CreateTemp(filepath.Dir(canonical), ".threadfin-rollback-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err = io.Copy(temporary, input); err != nil {
		_ = temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err = temporary.Chmod(0755); err != nil {
		_ = temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	if err = os.Remove(canonical); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(temporaryPath, canonical); err != nil {
		return err
	}
	return nil
}

func preparePrivateUpdateStartup(args []string) (privateUpdateStartup, error) {
	if len(args) < 2 || (args[1] != windowsUpdateHelperMode && args[1] != windowsUpdateChildMode) {
		return privateUpdateStartup{}, nil
	}
	result := privateUpdateStartup{private: true, exit: true, exitCode: 1}
	if runtime.GOOS != "windows" {
		return result, errors.New("private Windows update mode is unavailable on this platform")
	}
	protocol := defaultWindowsUpdateProtocol()
	switch args[1] {
	case windowsUpdateHelperMode:
		if len(args) != 4 {
			return result, errors.New("invalid private Windows update helper arguments")
		}
		if err := runWindowsUpdateHelper(args[2], args[3], protocol); err != nil {
			return result, err
		}
		result.exitCode = 0
		return result, nil
	case windowsUpdateChildMode:
		if len(args) != 6 {
			return result, errors.New("invalid private Windows update child arguments")
		}
		helperPID, err := strconv.Atoi(args[5])
		if err != nil {
			return result, errors.New("invalid private Windows update helper PID")
		}
		executable, err := protocol.executable()
		if err != nil {
			return result, err
		}
		child, err := loadWindowsUpdateChild(args[2], args[3], windowsUpdateAttempt(args[4]), helperPID, executable)
		if err != nil {
			return result, err
		}
		activeWindowsUpdateChild.Lock()
		if activeWindowsUpdateChild.child != nil {
			activeWindowsUpdateChild.Unlock()
			return result, errors.New("Windows update child mode was initialized twice")
		}
		activeWindowsUpdateChild.child = child
		activeWindowsUpdateChild.signaled = false
		activeWindowsUpdateChild.Unlock()
		result.exit = false
		result.exitCode = 0
		result.originalArgs = append([]string{}, child.originalArgs...)
		result.skipAutomaticUpdate = true
		return result, nil
	default:
		panic("unreachable private update mode")
	}
}

// PrepareUpdateStartup recognizes and executes the private Windows helper or
// restores the original argument vector for an update child.
func PrepareUpdateStartup(args []string) (private, exit bool, exitCode int, originalArgs []string, skipAutomaticUpdate bool, err error) {
	startup, err := preparePrivateUpdateStartup(args)
	return startup.private, startup.exit, startup.exitCode, startup.originalArgs, startup.skipAutomaticUpdate, err
}

// IsUpdateHandoff reports whether the old process must exit after transferring
// ownership to the acknowledged Windows helper.
func IsUpdateHandoff(err error) bool {
	return errors.Is(err, ErrWindowsUpdateHandoff)
}

// SignalUpdateReady marks an update child ready only after its HTTP listener is
// owned. Ordinary and non-Windows startups use this point for narrow cleanup.
func SignalUpdateReady() error {
	return signalWindowsUpdateReady()
}

func signalWindowsUpdateReady() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	activeWindowsUpdateChild.Lock()
	child := activeWindowsUpdateChild.child
	if child == nil {
		activeWindowsUpdateChild.Unlock()
		if err := retryCompletedWindowsUpdateCleanup(); err != nil {
			log.Printf("deferred Windows update cleanup retained recovery material: %v", err)
		}
		return nil
	}
	if activeWindowsUpdateChild.signaled {
		activeWindowsUpdateChild.Unlock()
		return errors.New("Windows update readiness was signaled more than once")
	}
	activeWindowsUpdateChild.signaled = true
	activeWindowsUpdateChild.Unlock()

	protocol := defaultWindowsUpdateProtocol()
	helper, err := protocol.findProcess(child.helperPID)
	if err != nil {
		return fmt.Errorf("open Windows update helper for cleanup: %w", err)
	}
	readyPath := windowsUpdateReadyPath(child.statePath, child.attempt)
	if err := writeWindowsUpdateMarker(readyPath, child.state.Nonce, windowsUpdateReadyMarker); err != nil {
		return errors.Join(err, helper.Release())
	}
	go func() {
		if err := finishWindowsUpdateChild(child.state, child.statePath, helper, protocol); err != nil {
			log.Printf("Windows update cleanup retained recovery material: %v", err)
		}
	}()
	return nil
}

func retryCompletedWindowsUpdateCleanup() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return retryCompletedWindowsUpdateCleanupForExecutable(executable)
}

func retryCompletedWindowsUpdateCleanupForExecutable(executable string) error {
	var err error
	executable, err = filepath.Abs(executable)
	if err != nil {
		return err
	}
	executable = filepath.Clean(executable)
	directory := filepath.Dir(executable)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	prefix := filepath.Base(executable) + ".update-"
	suffix := ".json"
	var result error
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || len(name) <= len(prefix)+len(suffix) {
			continue
		}
		if !sameUpdatePath(name[:len(prefix)], prefix) || !sameUpdatePath(name[len(name)-len(suffix):], suffix) {
			continue
		}
		nonce := name[len(prefix) : len(name)-len(suffix)]
		statePath := filepath.Join(directory, name)
		state, readErr := readWindowsUpdateState(statePath, nonce, executable)
		if readErr != nil {
			result = errors.Join(result, readErr)
			continue
		}
		complete, markerErr := windowsUpdateMarkerExists(windowsUpdateCompletionPath(statePath), nonce, windowsUpdateCompleteMarker)
		if markerErr != nil {
			result = errors.Join(result, markerErr)
			continue
		}
		if complete {
			result = errors.Join(result, cleanupCompletedWindowsUpdate(state, statePath))
		}
	}
	return result
}
