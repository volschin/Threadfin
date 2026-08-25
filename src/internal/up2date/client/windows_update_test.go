package up2date

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeUpdateProcess struct {
	pid          int
	label        string
	events       *[]string
	waitStarted  chan struct{}
	waitUntil    chan struct{}
	exited       bool
	exitedErr    error
	killErr      error
	waitErr      error
	releaseErr   error
	killCalls    int
	waitCalls    int
	releaseCalls int
}

func (process *fakeUpdateProcess) PID() int {
	return process.pid
}

func (process *fakeUpdateProcess) Exited() (bool, error) {
	return process.exited, process.exitedErr
}

func (process *fakeUpdateProcess) Kill() error {
	process.killCalls++
	process.record("kill")
	if process.killErr == nil {
		process.exited = true
	}
	return process.killErr
}

func (process *fakeUpdateProcess) Wait(time.Duration) error {
	process.waitCalls++
	process.record("wait")
	if process.waitStarted != nil {
		close(process.waitStarted)
	}
	if process.waitUntil != nil {
		<-process.waitUntil
	}
	return process.waitErr
}

func (process *fakeUpdateProcess) Release() error {
	process.releaseCalls++
	process.record("release")
	return process.releaseErr
}

func (process *fakeUpdateProcess) record(operation string) {
	if process.events != nil {
		*process.events = append(*process.events, process.label+":"+operation)
	}
}

func (process *fakeUpdateProcess) assertOwnedExactlyOnce(t *testing.T) {
	t.Helper()
	if process.waitCalls+process.releaseCalls != 1 {
		t.Fatalf("%s ownership: Wait=%d Release=%d, want exactly one", process.label, process.waitCalls, process.releaseCalls)
	}
	if process.killCalls > 0 && process.waitCalls != 1 {
		t.Fatalf("%s was killed %d time(s) but waited %d time(s)", process.label, process.killCalls, process.waitCalls)
	}
}

func immediateWindowsPoll(_ time.Duration, check func() (bool, error)) (bool, error) {
	for range 4 {
		done, err := check()
		if done || err != nil {
			return done, err
		}
	}
	return false, nil
}

func writeTestWindowsState(t *testing.T, args []string) (windowsUpdateState, string) {
	t.Helper()
	directory := t.TempDir()
	canonical := filepath.Join(directory, "Threadfin.exe")
	backup := filepath.Join(directory, "_old_Threadfin.exe")
	nonce := strings.Repeat("ab", windowsUpdateNonceBytes)
	state := windowsUpdateState{
		Version:   windowsUpdateStateVersion,
		Nonce:     nonce,
		Canonical: canonical,
		Backup:    backup,
		Args:      append([]string(nil), args...),
		OldPID:    41,
	}
	statePath := windowsUpdateStatePath(canonical, nonce)
	if err := writeWindowsUpdateState(statePath, state); err != nil {
		t.Fatal(err)
	}
	return state, statePath
}

func TestBeginWindowsHandoffAcknowledgesHelperBeforeOwnershipTransfer(t *testing.T) {
	directory := t.TempDir()
	canonical := filepath.Join(directory, "Threadfin.exe")
	backup := filepath.Join(directory, "_old_Threadfin.exe")
	originalArgs := []string{"-config", `C:\Threadfin Data\`, `--literal="quoted value"`}
	events := []string{}
	helper := &fakeUpdateProcess{pid: 84, label: "helper", events: &events}
	slowStartupBudget := 10 * time.Minute
	previousReadinessTimeout := Updater.WindowsUpdateReadinessTimeout
	Updater.WindowsUpdateReadinessTimeout = slowStartupBudget
	t.Cleanup(func() { Updater.WindowsUpdateReadinessTimeout = previousReadinessTimeout })
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 41 }
	protocol.pollCondition = immediateWindowsPoll
	protocol.startProcess = func(executable string, args []string) (updateProcess, error) {
		events = append(events, "start-helper")
		if executable != backup {
			t.Fatalf("helper executable = %q, want %q", executable, backup)
		}
		if len(args) != 3 || args[0] != windowsUpdateHelperMode {
			t.Fatalf("helper args = %#v", args)
		}
		state, err := readWindowsUpdateState(args[1], args[2], backup)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(state.Args, originalArgs) {
			t.Fatalf("saved args = %#v, want %#v", state.Args, originalArgs)
		}
		if state.ReadinessTimeout != slowStartupBudget {
			t.Fatalf("saved readiness budget = %v, want %v", state.ReadinessTimeout, slowStartupBudget)
		}
		if err := writeWindowsUpdateMarker(windowsUpdateAckPath(args[1]), args[2], windowsUpdateAckMarker); err != nil {
			t.Fatal(err)
		}
		return helper, nil
	}

	err := beginWindowsHandoff(canonical, backup, originalArgs, func() error {
		events = append(events, "cleanup-candidate")
		return nil
	}, protocol)
	if !errors.Is(err, ErrWindowsUpdateHandoff) {
		t.Fatalf("handoff error = %v, want ErrWindowsUpdateHandoff", err)
	}
	if !reflect.DeepEqual(events, []string{"cleanup-candidate", "start-helper", "helper:release"}) {
		t.Fatalf("events = %#v", events)
	}
	helper.assertOwnedExactlyOnce(t)
}

func TestBeginWindowsHandoffRejectsAcknowledgementFromExitedHelper(t *testing.T) {
	directory := t.TempDir()
	canonical := filepath.Join(directory, "Threadfin.exe")
	backup := filepath.Join(directory, "_old_Threadfin.exe")
	if err := os.WriteFile(canonical, []byte("replacement"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("known good"), 0755); err != nil {
		t.Fatal(err)
	}
	helper := &fakeUpdateProcess{pid: 84, label: "helper", exited: true}
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 41 }
	protocol.pollCondition = immediateWindowsPoll
	protocol.startProcess = func(_ string, args []string) (updateProcess, error) {
		if err := writeWindowsUpdateMarker(windowsUpdateAckPath(args[1]), args[2], windowsUpdateAckMarker); err != nil {
			t.Fatal(err)
		}
		return helper, nil
	}

	err := beginWindowsHandoff(canonical, backup, []string{"-port", "34400"}, func() error { return nil }, protocol)
	if errors.Is(err, ErrWindowsUpdateHandoff) {
		t.Fatal("ownership transferred to an already exited helper")
	}
	got, readErr := os.ReadFile(canonical)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "known good" {
		t.Fatalf("canonical executable = %q after dead-helper acknowledgement", got)
	}
	helper.assertOwnedExactlyOnce(t)
}

func TestWindowsHelperWaitsForOldExitAndPreservesOriginalArguments(t *testing.T) {
	originalArgs := []string{"-config", `C:\path with spaces\`, `--quote="a b"`, `trailing\\`}
	state, statePath := writeTestWindowsState(t, originalArgs)
	events := []string{}
	oldWaitStarted := make(chan struct{})
	oldExited := make(chan struct{})
	oldProcess := &fakeUpdateProcess{
		pid:         state.OldPID,
		label:       "old",
		events:      &events,
		waitStarted: oldWaitStarted,
		waitUntil:   oldExited,
	}
	replacement := &fakeUpdateProcess{pid: 85, label: "replacement", events: &events}
	type replacementStart struct {
		executable string
		args       []string
	}
	replacementStarted := make(chan replacementStart, 1)
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 84 }
	protocol.executable = func() (string, error) { return state.Backup, nil }
	protocol.findProcess = func(pid int) (updateProcess, error) {
		if pid != state.OldPID {
			t.Fatalf("find PID = %d, want %d", pid, state.OldPID)
		}
		return oldProcess, nil
	}
	protocol.startProcess = func(executable string, args []string) (updateProcess, error) {
		events = append(events, "start-replacement")
		replacementStarted <- replacementStart{executable: executable, args: append([]string(nil), args...)}
		return replacement, nil
	}
	protocol.pollCondition = func(_ time.Duration, check func() (bool, error)) (bool, error) {
		done, err := check()
		if done || err != nil {
			return done, err
		}
		if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsReplacementAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
			return false, err
		}
		return check()
	}

	helperDone := make(chan error, 1)
	go func() {
		helperDone <- runWindowsUpdateHelper(statePath, state.Nonce, protocol)
	}()
	select {
	case <-oldWaitStarted:
	case <-time.After(time.Second):
		t.Fatal("helper did not begin waiting for the old process")
	}
	select {
	case call := <-replacementStarted:
		t.Fatalf("replacement started before old exit: %#v", call)
	default:
	}
	close(oldExited)
	var call replacementStart
	select {
	case call = <-replacementStarted:
	case <-time.After(time.Second):
		t.Fatal("replacement did not start after old exit")
	}
	if call.executable != state.Canonical {
		t.Fatalf("replacement executable = %q, want %q", call.executable, state.Canonical)
	}
	want := []string{windowsUpdateChildMode, statePath, state.Nonce, string(windowsReplacementAttempt), "84"}
	if !reflect.DeepEqual(call.args, want) {
		t.Fatalf("replacement args = %#v, want %#v", call.args, want)
	}
	helperPID, err := strconv.Atoi(call.args[4])
	if err != nil {
		t.Fatal(err)
	}
	child, err := loadWindowsUpdateChild(call.args[1], call.args[2], windowsUpdateAttempt(call.args[3]), helperPID, call.executable)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(child.originalArgs, originalArgs) {
		t.Fatalf("restored args = %#v, want %#v", child.originalArgs, originalArgs)
	}
	select {
	case err := <-helperDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("helper did not finish after replacement readiness")
	}
	wantEvents := []string{"old:wait", "start-replacement", "replacement:release"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	oldProcess.assertOwnedExactlyOnce(t)
	replacement.assertOwnedExactlyOnce(t)
}

func TestWindowsReadinessRequiresMarkerWhileLiveChildIsReleased(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	child := &fakeUpdateProcess{pid: 85, label: "replacement"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.pollCondition = func(_ time.Duration, check func() (bool, error)) (bool, error) {
		done, err := check()
		if done || err != nil {
			t.Fatalf("live child became ready without listener marker: done=%t err=%v", done, err)
		}
		if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsReplacementAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
			t.Fatal(err)
		}
		return check()
	}

	if err := awaitWindowsUpdateReadiness(child, state, statePath, windowsReplacementAttempt, protocol); err != nil {
		t.Fatal(err)
	}
	if child.waitCalls != 0 || child.releaseCalls != 1 {
		t.Fatalf("live child ownership: Wait=%d Release=%d", child.waitCalls, child.releaseCalls)
	}
}

func TestWindowsReplacementUsesConfiguredSlowStartupBudget(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	state.ReadinessTimeout = 10 * time.Minute
	child := &fakeUpdateProcess{pid: 85, label: "replacement"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.pollCondition = func(timeout time.Duration, check func() (bool, error)) (bool, error) {
		if timeout != 10*time.Minute {
			t.Fatalf("replacement readiness budget = %v, want 10m", timeout)
		}
		if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsReplacementAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
			return false, err
		}
		return check()
	}

	if err := awaitWindowsUpdateReadiness(child, state, statePath, windowsReplacementAttempt, protocol); err != nil {
		t.Fatal(err)
	}
	child.assertOwnedExactlyOnce(t)
}

func TestWindowsReadinessRejectsExitedChildEvenWithMarker(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsReplacementAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
		t.Fatal(err)
	}
	child := &fakeUpdateProcess{pid: 85, label: "replacement", exited: true}
	protocol := defaultWindowsUpdateProtocol()
	protocol.pollCondition = immediateWindowsPoll

	if err := awaitWindowsUpdateReadiness(child, state, statePath, windowsReplacementAttempt, protocol); err == nil {
		t.Fatal("exited replacement was accepted as live listener owner")
	}
	if child.waitCalls != 1 || child.releaseCalls != 0 {
		t.Fatalf("exited child ownership: Wait=%d Release=%d", child.waitCalls, child.releaseCalls)
	}
	child.assertOwnedExactlyOnce(t)
}

func TestWindowsHelperRollsBackExitBeforeReadyAndStartsRecovery(t *testing.T) {
	originalArgs := []string{"-bind", "127.0.0.1", `C:\quoted path\`}
	state, statePath := writeTestWindowsState(t, originalArgs)
	if err := os.WriteFile(state.Canonical, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Backup, []byte("known good executable"), 0755); err != nil {
		t.Fatal(err)
	}
	oldProcess := &fakeUpdateProcess{pid: state.OldPID, label: "old"}
	replacement := &fakeUpdateProcess{pid: 85, label: "replacement", exited: true}
	recovery := &fakeUpdateProcess{pid: 86, label: "recovery"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 84 }
	protocol.executable = func() (string, error) { return state.Backup, nil }
	protocol.findProcess = func(int) (updateProcess, error) { return oldProcess, nil }
	startCalls := 0
	protocol.startProcess = func(executable string, args []string) (updateProcess, error) {
		startCalls++
		if executable != state.Canonical {
			t.Fatalf("child executable = %q, want %q", executable, state.Canonical)
		}
		if !reflect.DeepEqual(args[:3], []string{windowsUpdateChildMode, statePath, state.Nonce}) {
			t.Fatalf("child args prefix = %#v", args)
		}
		child, err := loadWindowsUpdateChild(statePath, state.Nonce, windowsUpdateAttempt(args[3]), 84, executable)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(child.originalArgs, originalArgs) {
			t.Fatalf("restored args = %#v, want %#v", child.originalArgs, originalArgs)
		}
		if startCalls == 1 {
			if args[3] != string(windowsReplacementAttempt) {
				t.Fatalf("first attempt = %q", args[3])
			}
			return replacement, nil
		}
		if args[3] != string(windowsRecoveryAttempt) {
			t.Fatalf("second attempt = %q", args[3])
		}
		return recovery, nil
	}
	pollCalls := 0
	protocol.pollCondition = func(_ time.Duration, check func() (bool, error)) (bool, error) {
		pollCalls++
		if pollCalls == 2 {
			if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsRecoveryAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
				t.Fatal(err)
			}
		}
		return check()
	}

	if err := runWindowsUpdateHelper(statePath, state.Nonce, protocol); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(state.Canonical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "known good executable" {
		t.Fatalf("restored executable = %q", got)
	}
	if _, err := os.Stat(state.Backup); err != nil {
		t.Fatalf("known-good helper was removed before recovery readiness: %v", err)
	}
	oldProcess.assertOwnedExactlyOnce(t)
	replacement.assertOwnedExactlyOnce(t)
	recovery.assertOwnedExactlyOnce(t)
}

func TestWindowsHelperKillsAndWaitsTimedOutReplacementBeforeRecovery(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-config", "data"})
	if err := os.WriteFile(state.Canonical, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Backup, []byte("known good executable"), 0755); err != nil {
		t.Fatal(err)
	}
	oldProcess := &fakeUpdateProcess{pid: state.OldPID, label: "old"}
	replacement := &fakeUpdateProcess{pid: 85, label: "replacement"}
	recovery := &fakeUpdateProcess{pid: 86, label: "recovery"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 84 }
	protocol.executable = func() (string, error) { return state.Backup, nil }
	protocol.findProcess = func(int) (updateProcess, error) { return oldProcess, nil }
	startCalls := 0
	protocol.startProcess = func(string, []string) (updateProcess, error) {
		startCalls++
		if startCalls == 1 {
			return replacement, nil
		}
		return recovery, nil
	}
	pollCalls := 0
	protocol.pollCondition = func(_ time.Duration, check func() (bool, error)) (bool, error) {
		pollCalls++
		if pollCalls == 1 {
			return false, nil
		}
		if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsRecoveryAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
			t.Fatal(err)
		}
		return check()
	}

	if err := runWindowsUpdateHelper(statePath, state.Nonce, protocol); err != nil {
		t.Fatal(err)
	}
	if replacement.killCalls != 1 || replacement.waitCalls != 1 || replacement.releaseCalls != 0 {
		t.Fatalf("timed-out replacement lifecycle: Kill=%d Wait=%d Release=%d", replacement.killCalls, replacement.waitCalls, replacement.releaseCalls)
	}
	replacement.assertOwnedExactlyOnce(t)
	recovery.assertOwnedExactlyOnce(t)
}

func TestWindowsRecoveryUsesConfiguredSlowStartupBudgetAfterRollback(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-config", "data"})
	state.ReadinessTimeout = 10 * time.Minute
	if err := os.Remove(statePath); err != nil {
		t.Fatal(err)
	}
	if err := writeWindowsUpdateState(statePath, state); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Canonical, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Backup, []byte("known good executable"), 0755); err != nil {
		t.Fatal(err)
	}
	oldProcess := &fakeUpdateProcess{pid: state.OldPID, label: "old"}
	replacement := &fakeUpdateProcess{pid: 85, label: "replacement"}
	recovery := &fakeUpdateProcess{pid: 86, label: "recovery"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 84 }
	protocol.executable = func() (string, error) { return state.Backup, nil }
	protocol.findProcess = func(int) (updateProcess, error) { return oldProcess, nil }
	startCalls := 0
	protocol.startProcess = func(string, []string) (updateProcess, error) {
		startCalls++
		if startCalls == 1 {
			return replacement, nil
		}
		return recovery, nil
	}
	pollCalls := 0
	protocol.pollCondition = func(timeout time.Duration, check func() (bool, error)) (bool, error) {
		pollCalls++
		if timeout != 10*time.Minute {
			t.Fatalf("attempt %d readiness budget = %v, want 10m", pollCalls, timeout)
		}
		if pollCalls == 1 {
			return false, nil
		}
		if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsRecoveryAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
			return false, err
		}
		return check()
	}

	if err := runWindowsUpdateHelper(statePath, state.Nonce, protocol); err != nil {
		t.Fatal(err)
	}
	if pollCalls != 2 {
		t.Fatalf("readiness polls = %d, want replacement and recovery", pollCalls)
	}
	if replacement.killCalls != 1 || replacement.waitCalls != 1 {
		t.Fatalf("replacement lifecycle: Kill=%d Wait=%d", replacement.killCalls, replacement.waitCalls)
	}
	oldProcess.assertOwnedExactlyOnce(t)
	replacement.assertOwnedExactlyOnce(t)
	recovery.assertOwnedExactlyOnce(t)
}

func TestWindowsHelperRetainsRecoveryMaterialWhenRestoredReadinessFails(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-config", "data"})
	if err := os.WriteFile(state.Canonical, []byte("new executable"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Backup, []byte("known good executable"), 0755); err != nil {
		t.Fatal(err)
	}
	oldProcess := &fakeUpdateProcess{pid: state.OldPID, label: "old"}
	replacement := &fakeUpdateProcess{pid: 85, label: "replacement", exited: true}
	recovery := &fakeUpdateProcess{pid: 86, label: "recovery", exited: true}
	protocol := defaultWindowsUpdateProtocol()
	protocol.currentPID = func() int { return 84 }
	protocol.executable = func() (string, error) { return state.Backup, nil }
	protocol.findProcess = func(int) (updateProcess, error) { return oldProcess, nil }
	startCalls := 0
	protocol.startProcess = func(string, []string) (updateProcess, error) {
		startCalls++
		if startCalls == 1 {
			return replacement, nil
		}
		return recovery, nil
	}
	protocol.pollCondition = immediateWindowsPoll

	if err := runWindowsUpdateHelper(statePath, state.Nonce, protocol); err == nil {
		t.Fatal("replacement and restored startup failures returned success")
	}
	for _, path := range []string{state.Backup, statePath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("recovery material %q was not retained: %v", path, err)
		}
	}
	got, err := os.ReadFile(state.Canonical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "known good executable" {
		t.Fatalf("canonical executable = %q, want known-good bytes", got)
	}
	oldProcess.assertOwnedExactlyOnce(t)
	replacement.assertOwnedExactlyOnce(t)
	recovery.assertOwnedExactlyOnce(t)
}

func TestFinishedWindowsChildWaitsForHelperThenCleansRecoveryMaterial(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	for path, body := range map[string]string{
		state.Canonical: "replacement",
		state.Backup:    "known good",
	} {
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeWindowsUpdateMarker(windowsUpdateAckPath(statePath), state.Nonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	}
	if err := writeWindowsUpdateMarker(windowsUpdateReadyPath(statePath, windowsReplacementAttempt), state.Nonce, windowsUpdateReadyMarker); err != nil {
		t.Fatal(err)
	}
	help := &fakeUpdateProcess{pid: 84, label: "helper"}

	if err := finishWindowsUpdateChild(state, statePath, help, defaultWindowsUpdateProtocol()); err != nil {
		t.Fatal(err)
	}
	help.assertOwnedExactlyOnce(t)
	for _, path := range []string{
		state.Backup,
		statePath,
		windowsUpdateAckPath(statePath),
		windowsUpdateReadyPath(statePath, windowsReplacementAttempt),
		windowsUpdateCompletionPath(statePath),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("update recovery material remains at %q: %v", path, err)
		}
	}
}

func TestFinishedWindowsChildPublishesCompletionAfterHelperWait(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	if err := os.Mkdir(state.Backup, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(state.Backup, "retain"), []byte("known good"), 0600); err != nil {
		t.Fatal(err)
	}
	helper := &fakeUpdateProcess{pid: 84, label: "helper"}

	if err := finishWindowsUpdateChild(state, statePath, helper, defaultWindowsUpdateProtocol()); err == nil {
		t.Fatal("injected cleanup failure returned success")
	}
	completePath := windowsUpdateCompletionPath(statePath)
	if complete, err := windowsUpdateMarkerExists(completePath, state.Nonce, windowsUpdateCompleteMarker); err != nil {
		t.Fatal(err)
	} else if !complete {
		t.Fatal("successful helper wait did not publish terminal completion before cleanup")
	}
	helper.assertOwnedExactlyOnce(t)
}

func TestHelperWaitTimeoutDoesNotAuthorizeDeferredCleanup(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	for path, body := range map[string]string{
		state.Canonical: "replacement",
		state.Backup:    "known good",
	} {
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeWindowsUpdateMarker(windowsUpdateAckPath(statePath), state.Nonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	}
	readyPath := windowsUpdateReadyPath(statePath, windowsReplacementAttempt)
	if err := writeWindowsUpdateMarker(readyPath, state.Nonce, windowsUpdateReadyMarker); err != nil {
		t.Fatal(err)
	}
	helper := &fakeUpdateProcess{pid: 84, label: "helper", waitErr: errWindowsUpdateTimeout}

	if err := finishWindowsUpdateChild(state, statePath, helper, defaultWindowsUpdateProtocol()); err == nil {
		t.Fatal("helper wait timeout was accepted as terminal completion")
	}
	completePath := windowsUpdateCompletionPath(statePath)
	if complete, err := windowsUpdateMarkerExists(completePath, state.Nonce, windowsUpdateCompleteMarker); err != nil {
		t.Fatal(err)
	} else if complete {
		t.Fatal("completion marker was published before helper exit was confirmed")
	}
	if err := retryCompletedWindowsUpdateCleanupForExecutable(state.Canonical); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		state.Canonical,
		state.Backup,
		statePath,
		windowsUpdateAckPath(statePath),
		readyPath,
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("recovery material %q was removed without terminal completion: %v", path, err)
		}
	}
	helper.assertOwnedExactlyOnce(t)
}

func TestFinishedWindowsChildRetriesFailOnceCompletionPublication(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	for path, body := range map[string]string{
		state.Canonical: "replacement",
		state.Backup:    "known good",
	} {
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	helper := &fakeUpdateProcess{pid: 84, label: "helper"}
	protocol := defaultWindowsUpdateProtocol()
	protocol.pollCondition = immediateWindowsPoll
	publicationAttempts := 0
	protocol.writeMarker = func(path, nonce, kind string) error {
		publicationAttempts++
		if path != windowsUpdateCompletionPath(statePath) || nonce != state.Nonce || kind != windowsUpdateCompleteMarker {
			return errors.New("completion publication received wrong marker identity")
		}
		if publicationAttempts == 1 {
			return errors.New("injected transient completion failure")
		}
		return writeWindowsUpdateMarker(path, nonce, kind)
	}

	if err := finishWindowsUpdateChild(state, statePath, helper, protocol); err != nil {
		t.Fatal(err)
	}
	if publicationAttempts != 2 {
		t.Fatalf("completion publication attempts = %d, want 2", publicationAttempts)
	}
	helper.assertOwnedExactlyOnce(t)
}

func TestWindowsUpdateMarkerPublicationIsAbsentUntilComplete(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "marker.ack")
	nonce := strings.Repeat("ab", windowsUpdateNonceBytes)
	prePublication := make(chan struct{})
	allowPublication := make(chan struct{})
	ops := defaultWindowsUpdateMarkerOps()
	ops.beforePublish = func() {
		close(prePublication)
		<-allowPublication
	}
	writeDone := make(chan error, 1)
	go func() {
		writeDone <- writeWindowsUpdateMarkerWithOps(path, nonce, windowsUpdateAckMarker, ops)
	}()

	select {
	case <-prePublication:
	case <-time.After(time.Second):
		t.Fatal("marker writer did not reach the pre-publication barrier")
	}
	if exists, err := windowsUpdateMarkerExists(path, nonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	} else if exists {
		t.Fatal("final marker path was visible before publication")
	}
	close(allowPublication)
	select {
	case err := <-writeDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("marker publication did not finish")
	}
	if exists, err := windowsUpdateMarkerExists(path, nonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	} else if !exists {
		t.Fatal("published marker was not complete")
	}
	temporary, err := filepath.Glob(filepath.Join(directory, ".marker.ack.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporary) != 0 {
		t.Fatalf("temporary marker files remain after publication: %#v", temporary)
	}
}

func TestWindowsUpdateMarkerPublicationFailureNeverAdvancesProtocol(t *testing.T) {
	nonce := strings.Repeat("ab", windowsUpdateNonceBytes)
	failure := errors.New("injected marker publication failure")
	tests := []struct {
		name   string
		mutate func(*windowsUpdateMarkerOps)
	}{
		{
			name: "close",
			mutate: func(ops *windowsUpdateMarkerOps) {
				ops.closeFile = func(file *os.File) error {
					if err := file.Close(); err != nil {
						return err
					}
					return failure
				}
			},
		},
		{
			name: "publish",
			mutate: func(ops *windowsUpdateMarkerOps) {
				ops.publish = func(string, string) error { return failure }
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "marker.ready")
			ops := defaultWindowsUpdateMarkerOps()
			test.mutate(&ops)
			if err := writeWindowsUpdateMarkerWithOps(path, nonce, windowsUpdateReadyMarker, ops); !errors.Is(err, failure) {
				t.Fatalf("write error = %v, want injected failure", err)
			}
			if exists, err := windowsUpdateMarkerExists(path, nonce, windowsUpdateReadyMarker); err != nil {
				t.Fatal(err)
			} else if exists {
				t.Fatal("protocol advanced after incomplete marker publication")
			}
			temporary, err := filepath.Glob(filepath.Join(directory, ".marker.ready.tmp-*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(temporary) != 0 {
				t.Fatalf("temporary marker files remain after failure: %#v", temporary)
			}
		})
	}
}

func TestWindowsUpdateMarkerPublicationNeverReplacesExistingMarker(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "marker.ack")
	firstNonce := strings.Repeat("ab", windowsUpdateNonceBytes)
	secondNonce := strings.Repeat("cd", windowsUpdateNonceBytes)
	if err := writeWindowsUpdateMarker(path, firstNonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	}
	if err := writeWindowsUpdateMarker(path, secondNonce, windowsUpdateAckMarker); err == nil {
		t.Fatal("second publisher replaced an existing marker")
	}
	if exists, err := windowsUpdateMarkerExists(path, firstNonce, windowsUpdateAckMarker); err != nil {
		t.Fatal(err)
	} else if !exists {
		t.Fatal("first publisher's complete marker was replaced")
	}
}

func TestDeferredWindowsUpdateCleanupDiscoversStateCaseInsensitively(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-port", "34400"})
	for path, body := range map[string]string{
		state.Canonical: "replacement",
		state.Backup:    "known good",
	} {
		if err := os.WriteFile(path, []byte(body), 0755); err != nil {
			t.Fatal(err)
		}
	}
	casedStatePath := filepath.Join(
		filepath.Dir(statePath),
		"tHrEaDfIn.ExE.UpDaTe-"+state.Nonce+".JsOn",
	)
	if err := os.Rename(statePath, casedStatePath); err != nil {
		t.Fatal(err)
	}
	if err := writeWindowsUpdateMarker(windowsUpdateCompletionPath(casedStatePath), state.Nonce, windowsUpdateCompleteMarker); err != nil {
		t.Fatal(err)
	}
	casedExecutable := filepath.Join(filepath.Dir(state.Canonical), "THREADFIN.EXE")

	if err := retryCompletedWindowsUpdateCleanupForExecutable(casedExecutable); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		state.Backup,
		casedStatePath,
		windowsUpdateCompletionPath(casedStatePath),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("case-insensitive cleanup left %q: %v", path, err)
		}
	}
	if _, err := os.Stat(state.Canonical); err != nil {
		t.Fatalf("canonical executable was removed: %v", err)
	}
}

func TestWindowsUpdateStateRejectsInvalidPathAndNonceRelationships(t *testing.T) {
	state, statePath := writeTestWindowsState(t, []string{"-config", "data"})
	tests := []struct {
		name   string
		mutate func(*windowsUpdateState, *string, *string, *string)
	}{
		{
			name: "relative canonical",
			mutate: func(candidate *windowsUpdateState, _, _, _ *string) {
				candidate.Canonical = "Threadfin.exe"
			},
		},
		{
			name: "unrelated backup",
			mutate: func(candidate *windowsUpdateState, _, _, _ *string) {
				candidate.Backup = filepath.Join(filepath.Dir(candidate.Canonical), "other.exe")
			},
		},
		{
			name: "state outside canonical relationship",
			mutate: func(_ *windowsUpdateState, path, _, _ *string) {
				*path = filepath.Join(filepath.Dir(*path), "other-state.json")
			},
		},
		{
			name: "nonce mismatch",
			mutate: func(_ *windowsUpdateState, _, nonce, _ *string) {
				*nonce = strings.Repeat("cd", windowsUpdateNonceBytes)
			},
		},
		{
			name: "malformed nonce",
			mutate: func(candidate *windowsUpdateState, path, nonce, _ *string) {
				candidate.Nonce = "not-random"
				*nonce = candidate.Nonce
				*path = windowsUpdateStatePath(candidate.Canonical, candidate.Nonce)
			},
		},
		{
			name: "wrong executable generation",
			mutate: func(_ *windowsUpdateState, _, _, executable *string) {
				*executable = filepath.Join(filepath.Dir(*executable), "attacker.exe")
			},
		},
		{
			name: "readiness budget below protocol minimum",
			mutate: func(candidate *windowsUpdateState, _, _, _ *string) {
				candidate.ReadinessTimeout = time.Second
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := state
			candidate.Args = append([]string(nil), state.Args...)
			candidatePath := statePath
			nonce := state.Nonce
			executable := state.Backup
			test.mutate(&candidate, &candidatePath, &nonce, &executable)
			if err := validateWindowsUpdateState(candidate, candidatePath, nonce, executable); err == nil {
				t.Fatal("invalid helper/child state was accepted")
			}
		})
	}
}
