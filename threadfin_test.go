package main

import (
	"errors"
	"os"
	"reflect"
	"regexp"
	"testing"

	"threadfin/src"
)

func TestDockerfileNvidiaBaseMatchesJellyfinRepository(t *testing.T) {
	dockerfile, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatal(err)
	}

	wantUbuntuByCodename := map[string]string{
		"noble":    "24.04",
		"resolute": "26.04",
	}
	baseStages := regexp.MustCompile(`(?m)^FROM (?:nvidia/cuda:[^[:space:]]+-ubuntu|ubuntu:)([0-9]+\.[0-9]+) AS (standard|nvidia)\nENV OS_CODENAME=([^[:space:]]+)$`).FindAllSubmatch(dockerfile, -1)
	if len(baseStages) != 2 {
		t.Fatalf("Dockerfile has %d codename-mapped final base stages, want 2", len(baseStages))
	}
	for _, stage := range baseStages {
		ubuntuVersion := string(stage[1])
		stageName := string(stage[2])
		jellyfinCodename := string(stage[3])
		wantUbuntu, ok := wantUbuntuByCodename[jellyfinCodename]
		if !ok {
			t.Fatalf("stage %q uses unsupported Jellyfin Ubuntu codename %q", stageName, jellyfinCodename)
		}
		if ubuntuVersion != wantUbuntu {
			t.Fatalf("stage %q Ubuntu version = %q, want %q for Jellyfin codename %q", stageName, ubuntuVersion, wantUbuntu, jellyfinCodename)
		}
	}
}

func TestDispatchThreadfinStartupHandlesPrivateModeBeforeApplication(t *testing.T) {
	applicationCalled := false
	exitCode := dispatchThreadfinStartup(
		[]string{"Threadfin.exe", "--private-mode-that-normal-flags-reject"},
		func([]string) (src.UpdateStartup, error) {
			return src.UpdateStartup{Private: true, Exit: true, ExitCode: 7}, nil
		},
		func([]string, bool) int {
			applicationCalled = true
			return 0
		},
	)
	if exitCode != 7 {
		t.Fatalf("exit code = %d, want 7", exitCode)
	}
	if applicationCalled {
		t.Fatal("ordinary application and flag parsing ran for private helper mode")
	}
}

func TestDispatchThreadfinStartupRestoresChildArgumentsAndSkipsOneUpdate(t *testing.T) {
	originalArgs := []string{"-config", `C:\path with spaces\`, `--quoted="a b"`, `trailing\\`}
	gotArgs := []string(nil)
	gotSkip := false
	exitCode := dispatchThreadfinStartup(
		[]string{"Threadfin.exe", "--private-child", "state"},
		func([]string) (src.UpdateStartup, error) {
			return src.UpdateStartup{
				Private:             true,
				OriginalArgs:        originalArgs,
				SkipAutomaticUpdate: true,
			}, nil
		},
		func(args []string, skipAutomaticUpdate bool) int {
			gotArgs = append([]string(nil), args...)
			gotSkip = skipAutomaticUpdate
			return 0
		},
	)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	wantArgs := append([]string{"Threadfin.exe"}, originalArgs...)
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("application args = %#v, want %#v", gotArgs, wantArgs)
	}
	if !gotSkip {
		t.Fatal("update child did not skip the automatic update that caused its restart")
	}
}

func TestDispatchThreadfinStartupReturnsNonzeroForPrivateModeFailure(t *testing.T) {
	exitCode := dispatchThreadfinStartup(
		[]string{"Threadfin.exe", "--private-helper"},
		func([]string) (src.UpdateStartup, error) {
			return src.UpdateStartup{Private: true, Exit: true, ExitCode: 1}, errors.New("invalid private state")
		},
		func([]string, bool) int {
			t.Fatal("ordinary application ran after private startup failure")
			return 0
		},
	)
	if exitCode == 0 {
		t.Fatal("private startup failure returned success")
	}
}

func TestPerformStartupUpdateSkipsOnlyUpdateChildCheck(t *testing.T) {
	src.System.ConfigurationWizard = false
	t.Cleanup(func() { src.System.ConfigurationWizard = false })
	calls := 0
	update := func() error {
		calls++
		return nil
	}
	if err := performStartupUpdate(true, update); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("child update calls = %d, want 0", calls)
	}
	if err := performStartupUpdate(false, update); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("ordinary update calls = %d, want 1", calls)
	}
}

func TestPerformStartupUpdateSkipsUnconfiguredFirstRun(t *testing.T) {
	src.System.ConfigurationWizard = true
	t.Cleanup(func() { src.System.ConfigurationWizard = false })

	calls := 0
	if err := performStartupUpdate(false, func() error {
		calls++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("first-run update calls = %d, want 0", calls)
	}
}
